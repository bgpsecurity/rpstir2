package chainvalidate

import (
	"errors"
	"strings"
	"sync"
	"time"

	belogs "github.com/astaxie/beego/logs"
	asnutil "github.com/cpusoft/goutil/asnutil"
	certutil "github.com/cpusoft/goutil/certutil"
	conf "github.com/cpusoft/goutil/conf"
	convert "github.com/cpusoft/goutil/convert"
	iputil "github.com/cpusoft/goutil/iputil"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"

	"chainvalidate/db"
	chainmodel "chainvalidate/model"
	"model"
)

func GetChainCers(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	belogs.Debug("GetChainCers(): start:")

	chainCerSqls, err := db.GetChainCerSqls()
	if err != nil {
		belogs.Error("GetChainCers(): db.GetChainCerSqls:", err)
		return
	}
	belogs.Debug("GetChainCers(): GetChainCers, len(chainCerSqls):", len(chainCerSqls))

	for i := range chainCerSqls {
		chainCer := chainCerSqls[i].ToChainCer()
		belogs.Debug("GetChainCers():i, chainCer:", i, jsonutil.MarshalJson(chainCer))
		chains.CerIds = append(chains.CerIds, chainCerSqls[i].Id)
		chains.AddCer(&chainCer)
	}

	belogs.Debug("GetChainCers(): end, len(chainCerSqls):", len(chainCerSqls), ",   len(chains.CerIds):", len(chains.CerIds), ",  time(s):", time.Now().Sub(start).Seconds())
	return
}

func ValidateCers(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	cerIds := chains.CerIds
	belogs.Debug("ValidateCers(): start: len(cerIds):", len(cerIds))

	var cerWg sync.WaitGroup
	chainCerCh := make(chan int, conf.Int("chain::chainConcurrentCount"))
	for _, cerId := range cerIds {
		cerWg.Add(1)
		chainCerCh <- 1
		go validateCer(chains, cerId, &cerWg, chainCerCh)
	}
	cerWg.Wait()
	close(chainCerCh)

	belogs.Info("ValidateCers():end len(cerIds):", len(cerIds), "  time(s):", time.Now().Sub(start).Seconds())
}

func validateCer(chains *chainmodel.Chains, cerId uint64, wg *sync.WaitGroup, chainCerCh chan int) {
	defer func() {
		wg.Done()
		<-chainCerCh
	}()

	start := time.Now()
	chainCer, err := chains.GetCerById(cerId)
	if err != nil {
		belogs.Error("validateCer(): cerId fail:", cerId, err)
		return
	}

	chainCer.ParentChainCerAlones, err = GetCerParentChainCers(chains, cerId)
	if err != nil {
		belogs.Error("validateCer(): GetCerParentChainCers fail:", cerId, err)
		chainCer.StateModel.JudgeState()
		chains.UpdateFileTypeIdToCer(&chainCer)
		return
	}
	belogs.Debug("validateCer():chainCer.ParentChainCers, cerId, len(chainCer.ParentChainCers):", cerId, len(chainCer.ParentChainCerAlones))

	chainCer.ChildChainCerAlones, chainCer.ChildChainCrls,
		chainCer.ChildChainMfts, chainCer.ChildChainRoas, err = getChildChainCersCrlsMftsRoas(chains, cerId)
	if err != nil {
		belogs.Error("validateCer(): getChildChainCersCrlsMftsRoas fail:", cerId, err)
		chainCer.StateModel.JudgeState()
		chains.UpdateFileTypeIdToCer(&chainCer)
		return
	}
	belogs.Debug("validateCer():chainCer.ChildChains, cerId:", cerId, len(chainCer.ChildChainCerAlones),
		len(chainCer.ChildChainCrls), len(chainCer.ChildChainMfts), len(chainCer.ChildChainRoas))

	// if is root cer, then verify self
	if chainCer.IsRoot {
		result, err := certutil.VerifyRootCerByOpenssl(osutil.JoinPathFile(chainCer.FilePath, chainCer.FileName))
		belogs.Debug("validateCer(): IsRoot VerifyRootCerByOpenssl result:", result, err)

		desc := ""
		if err != nil {
			desc = err.Error()
			belogs.Error("validateCer(): VerifyRootCerByOpenssl fail, fileName:", chainCer.FileName, result, err)
		}
		if result != "ok" {
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to self verification of root certificate",
				Detail: desc}
			chainCer.StateModel.AddError(&stateMsg)
		}
	} else {

		// if not root cer, should have parent cer
		if len(chainCer.ParentChainCerAlones) > 0 {
			// get one parent
			parentCer := osutil.JoinPathFile(chainCer.ParentChainCerAlones[0].FilePath, chainCer.ParentChainCerAlones[0].FileName)
			cer := osutil.JoinPathFile(chainCer.FilePath, chainCer.FileName)
			belogs.Debug("validateCer(): parentCer:", parentCer, "    cer:", cer)

			// openssl verify parent --> child
			result, err := certutil.VerifyCerByX509(parentCer, cer)
			belogs.Debug("validateCer(): VerifyCerByX509 parentCer:", parentCer, "   cer:", cer, "   result:", result, err)
			if result != "ok" {
				desc := ""
				if err != nil {
					desc = err.Error()
					belogs.Error("validateCer(): VerifyCerByX509 fail, cerId:", chainCer.Id, result, err)
				}
				stateMsg := model.StateMsg{Stage: "chainvalidate",
					Fail: "Fail to be verified by its issuing certificate",
					Detail: desc + ",  parent cer file is " + chainCer.ParentChainCerAlones[0].FileName +
						",  this cer file is " + chainCer.FileName}
				// if subject doesnot match ,will just set warning
				if strings.Contains(desc, "issuer name does not match subject from issuing certificate") {
					chainCer.StateModel.AddWarning(&stateMsg)
				} else {
					chainCer.StateModel.AddError(&stateMsg)
				}
			}

			// verify ipaddress prefix,if one parent is not found ,found the upper
			// rfc8360: Validation Reconsidered, set warning
			invalidIps := IpAddressesIncludeInParents(chainCer.ParentChainCerAlones, chainCer.ChainIpAddresses)
			if len(invalidIps) > 0 {
				belogs.Debug("validateCer(): cer ipaddress is overclaimed, fail, cerId:", chainCer.Id, jsonutil.MarshalJson(invalidIps), err)
				stateMsg := model.StateMsg{Stage: "chainvalidate",
					Fail:   "Certificate has overclaimed IP address not contained on the issuing certificate",
					Detail: "invalid ip are " + jsonutil.MarshalJson(invalidIps)}
				chainCer.StateModel.AddWarning(&stateMsg)
			}

			// verify ipaddress prefix,if one parent is not found ,found the upper
			// rfc8360: Validation Reconsidered, set warning
			invalidAsns := AsnsIncludeInParents(chainCer.ParentChainCerAlones, chainCer.ChainAsns)
			if len(invalidAsns) > 0 {
				belogs.Debug("validateCer(): cer asn is overclaimed, fail, cerId:", chainCer.Id, jsonutil.MarshalJson(invalidAsns), err)
				stateMsg := model.StateMsg{Stage: "chainvalidate",
					Fail:   "Certificate has overclaimed ASN not contained on the issuing certificate",
					Detail: "invalid asns are " + jsonutil.MarshalJson(invalidAsns)}
				chainCer.StateModel.AddWarning(&stateMsg)
			}

		} else {
			belogs.Debug("validateCer(): cer file has not found parent cer, fail,  chainCer.Id, cerId:", chainCer.Id, cerId)
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Its issuing certificate no longer exists",
				Detail: ""}
			chainCer.StateModel.AddError(&stateMsg)

		}
	}

	// check : must have one or mor mft and one crl child files
	if len(chainCer.ChildChainCrls) == 0 {
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Certificate does not issue at least one CRL",
			Detail: ""}
		chainCer.StateModel.AddError(&stateMsg)

	} /* else if len(chainCer.ChildChainCrls) > 1 {
		belogs.Debug("validateCer(): cer file find tow or more child crl files:",
			chainCer.Id, len(chainCer.ChildChainCrls))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "cer file find two or more child crl files",
			Detail: chainCer.FileName + " found " + convert.ToString(len(chainCer.ChildChainCrls)) + " crl files"}
		chainCer.StateModel.AddError(&stateMsg)
	}	*/

	if len(chainCer.ChildChainMfts) == 0 {
		belogs.Debug("validateCer(): cer file cannot find child mft file:", chainCer.Id)
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Certificate does not issue at least one Manifest",
			Detail: ""}
		if conf.Bool("policy::allowNoMft") {
			chainCer.StateModel.AddWarning(&stateMsg)
		} else {
			chainCer.StateModel.AddError(&stateMsg)
		}
	} /* else if len(chainCer.ChildChainMfts) > 1 {
		belogs.Debug("validateCer(): cer file find tow or more child mft files:",
			chainCer.Id, len(chainCer.ChildChainMfts))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "cer file find two or more child mft files",
			Detail: chainCer.FileName + " found " + convert.ToString(len(chainCer.ChildChainMfts)) + " mft files"}
		chainCer.StateModel.AddError(&stateMsg)
	}*/

	if len(chainCer.ChainSnInCrlRevoked.CrlFileName) > 0 {
		belogs.Debug("validateCer(): cer file is founded in crl's revoked cer list:",
			chainCer.Id, jsonutil.MarshalJson(chainCer.ChainSnInCrlRevoked.CrlFileName))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail: "Certificate is found on revocation list of CRL",
			Detail: chainCer.FileName + " is in " + chainCer.ChainSnInCrlRevoked.CrlFileName + " revoked cer list, " +
				" and revoked time is " + convert.Time2StringZone(chainCer.ChainSnInCrlRevoked.RevocationTime)}
		chainCer.StateModel.AddError(&stateMsg)
	}

	chainCer.StateModel.JudgeState()
	belogs.Debug("validateCer(): stateModel:", chainCer.StateModel)
	if chainCer.StateModel.State != "valid" {
		belogs.Info("validateCer(): stateModel have errors or warnings, cerId :", cerId, "  stateModel:", jsonutil.MarshalJson(chainCer.StateModel))
	}

	chains.UpdateFileTypeIdToCer(&chainCer)
	belogs.Debug("validateCer():end  UpdateFileTypeIdToCer: cerId:", cerId, "  time(s):", time.Now().Sub(start).Seconds())
	return

}

//
func AsnsIncludeInParents(parentChainCerAlones []chainmodel.ChainCerAlone, self []chainmodel.ChainAsn) (invalids []chainmodel.ChainAsn) {
	// self is inherit,there is no asn, then is ok
	if len(self) == 0 {
		return nil
	}

	// found ip one by one,
	for i := range parentChainCerAlones {
		if len(parentChainCerAlones[i].ChainAsns) == 0 {
			belogs.Debug("IncludeInParents(): len(parentChainCerAlones[i].ChainAsn) is 0 ")
			continue
		}

		invalids = asnsIncludeInParent(parentChainCerAlones[i].ChainAsns, self)
		if len(invalids) > 0 {
			belogs.Debug("IncludeInParents(): self asns:", jsonutil.MarshalJson(self),
				"   invalids:", jsonutil.MarshalJson(invalids))
			break
		}
	}
	return invalids
}

//
func asnsIncludeInParent(parents []chainmodel.ChainAsn, self []chainmodel.ChainAsn) (invalids []chainmodel.ChainAsn) {
	invalids = make([]chainmodel.ChainAsn, 0)
	for _, s := range self {
		// self is inherit, all are zero ,then is ok
		if s.Asn == 0 && s.Min == 0 && s.Max == 0 {
			continue
		}

		include := false
		for _, p := range parents {
			if p.Asn == 0 && p.Min == 0 && p.Max == 0 {
				continue
			}
			include = asnutil.AsnIncludeInParentAsn(s.Asn, s.Min, s.Max, p.Asn, p.Min, p.Max)
			if include {
				break
			}
		}
		if !include {
			belogs.Debug("asnsIncludeInParent():is not include: self:[",
				jsonutil.MarshalJson(self), "], parent:[", jsonutil.MarshalJson(parents), "]")
			invalids = append(invalids, s)
		}
	}
	return invalids
}

//parentss, self
func IpAddressesIncludeInParents(parentChainCerAlones []chainmodel.ChainCerAlone, self []chainmodel.ChainIpAddress) (invalids []chainmodel.ChainIpAddress) {
	// self is inherit ,then is ok
	if len(self) == 0 {
		belogs.Debug("IpAddressesIncludeInParents(): len(self) is 0 ")
		return nil
	}

	// found ip one by one,
	for i := range parentChainCerAlones {
		if len(parentChainCerAlones[i].ChainIpAddresses) == 0 {
			belogs.Debug("IpAddressesIncludeInParents(): len(parentChainCerAlones[i].ChainIpAddresses) is 0 ")
			continue
		}
		invalids = ipAddressesIncludeInParent(parentChainCerAlones[i].ChainIpAddresses, self)
		if len(invalids) > 0 {
			belogs.Debug("IpAddressesIncludeInParents(): found invalids, self ipaddress:", jsonutil.MarshalJson(self),
				"   invalids:", jsonutil.MarshalJson(invalids))
			break
		}
	}
	return invalids
}

// parents, self
func ipAddressesIncludeInParent(parents []chainmodel.ChainIpAddress, self []chainmodel.ChainIpAddress) (invalids []chainmodel.ChainIpAddress) {
	invalids = make([]chainmodel.ChainIpAddress, 0)
	for _, s := range self {
		include := false
		for _, p := range parents {
			include = iputil.IpRangeIncludeInParentRange(p.RangeStart, p.RangeEnd, s.RangeStart, s.RangeEnd)
			if include {
				belogs.Debug("ipAddressesIncludeInParent():is include: parent:[", p.RangeStart, p.RangeEnd,
					"],  self:[", s.RangeStart, s.RangeEnd, "]")
				break
			}
		}
		if !include {
			invalids = append(invalids, s)
			belogs.Debug("ipAddressesIncludeInParent():is not include: parents:", parents,
				"  self:[", s.RangeStart, s.RangeEnd, "]  invalids:", invalids)
		}
	}
	belogs.Debug("ipAddressesIncludeInParent():invalids:", invalids)
	return invalids
}

func GetCerParentChainCers(chains *chainmodel.Chains, cerId uint64) ([]chainmodel.ChainCerAlone, error) {

	chainCer, err := chains.GetCerById(cerId)
	if err != nil {
		belogs.Error("GetCerParentChainCers(): GetCer cerId:", cerId, err)
		return nil, err
	}
	belogs.Debug("GetCerParentChainCers(): cerId:", cerId, "  chainCer.Id:", chainCer.Id)

	chainCerAlones := make([]chainmodel.ChainCerAlone, 0, 10)

	// if is root, then just return
	if chainCer.IsRoot {
		belogs.Debug("GetCerParentChainCers(): cerId chainCer.IsRoot:", cerId, chainCer.IsRoot)
		return chainCerAlones, nil
	}

	for {

		parentChainCer, err := getCerParentChainCer(chains, cerId)
		belogs.Debug("GetCerParentChainCers(): cerId parentChainCer.Id, err:", cerId, parentChainCer.Id, err)
		if err != nil {
			belogs.Error("GetCerParentChainCers(): GetCerParentChainCer, cerId:", cerId, err)
			return nil, err
		}
		// not parent
		if parentChainCer.Id == 0 {
			belogs.Debug("GetCerParentChainCers(): GetCerParentChainCer,not found parent cer, cerId:", cerId)
			return chainCerAlones, nil
		}
		chainCerAlone := chainmodel.NewChainCerAlone(&parentChainCer)
		chainCerAlones = append(chainCerAlones, *chainCerAlone)
		belogs.Debug("GetCerParentChainCers(): cerId, len(chainCerAlones), added fileName:", cerId, len(chainCerAlones), chainCerAlone.FileName)
		if parentChainCer.IsRoot {
			belogs.Debug("GetCerParentChainCers(): IsRoot, cerId parentChainCer.Id :", cerId, parentChainCer.Id)
			return chainCerAlones, nil
		}
		cerId = parentChainCer.Id
	}
	belogs.Debug("GetCerParentChainCers(): cerId, len(chainCerAlones):", cerId, len(chainCerAlones))
	return chainCerAlones, nil
}
func getCerParentChainCer(chains *chainmodel.Chains, cerId uint64) (parentChainCer chainmodel.ChainCer, err error) {
	chainCer, err := chains.GetCerById(cerId)
	if err != nil {
		belogs.Error("getCerParentChainCer(): GetCerById, cerId:", cerId, err)
		return parentChainCer, err
	}
	belogs.Debug("getCerParentChainCer(): cerId:", cerId, "  chainCer.id", chainCer.Id)
	if chainCer.IsRoot {
		belogs.Debug("getCerParentChainCer(): GetCer  is root, cerId:", cerId, " chainCer.id:", chainCer.Id)
		return parentChainCer, nil
	}

	//get mft's aki --> parent cer's ski
	if len(chainCer.Aki) == 0 {
		belogs.Error("getCerParentChainCer(): chainCer.Aki is empty, fail:", cerId)
		return parentChainCer, errors.New("cer's aki is empty")
	}
	aki := chainCer.Aki
	parentCerSki := aki
	fileTypeId, ok := chains.SkiToFileTypeId[parentCerSki]
	belogs.Debug("getCerParentChainCer(): cerId,parentCerSki,fileTypeId, ok:", cerId, parentCerSki, fileTypeId, ok)
	if ok {
		parentChainCer, err = chains.GetCerByFileTypeId(fileTypeId)
		belogs.Debug("getCerParentChainCer(): GetCerByFileTypeId, cerId, fileTypeId, parentChainCer.Id:", cerId, fileTypeId, parentChainCer.Id)
		if err != nil {
			belogs.Error("getCerParentChainCer(): GetCerByFileTypeId, cerId,fileTypeId, fail:", cerId, fileTypeId, err)
			return parentChainCer, err
		}
		return parentChainCer, nil
	}
	//  not found parent ,is not error
	belogs.Debug("getCerParentChainCer(): not found cer's parent cer:", cerId)
	return parentChainCer, nil
}

func getChildChainCersCrlsMftsRoas(chains *chainmodel.Chains, cerId uint64) (childChainCerAlones []chainmodel.ChainCerAlone,
	childChainCrls []chainmodel.ChainCrl,
	childChainMfts []chainmodel.ChainMft,
	childChainRoas []chainmodel.ChainRoa, err error) {
	start := time.Now()

	chainCer, err := chains.GetCerById(cerId)
	if err != nil {
		belogs.Error("getChildChainCersCrlsMftsRoas(): GetCer, cerId:", cerId, err)
		return nil, nil, nil, nil, err
	}
	childChainCerAlones = make([]chainmodel.ChainCerAlone, 0)
	childChainCrls = make([]chainmodel.ChainCrl, 0)
	childChainMfts = make([]chainmodel.ChainMft, 0)
	childChainRoas = make([]chainmodel.ChainRoa, 0)

	// cer's ski --> child's aki
	ski := chainCer.Ski
	childsAki := ski
	fileTypeIds, ok := chains.AkiToFileTypeIds[childsAki]
	belogs.Debug("getChildChainCersCrlsMftsRoas(): cerId fileTypeIds, ok:", cerId, fileTypeIds, ok)
	if ok {
		for i := range fileTypeIds.FileTypeIds {
			fileTypeId := fileTypeIds.FileTypeIds[i]
			belogs.Debug("getChildChainCersCrlsMftsRoas(): cerId, fileTypeId, ok:", cerId, fileTypeId, ok)
			if ok {
				fileType := string(fileTypeId[:3])
				belogs.Debug("getChildChainCersCrlsMftsRoas(): cerId fileType:", cerId, fileType)
				switch fileType {
				case "cer":
					chainCerTmp, err := chains.GetCerByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getChildChainCersCrlsMftsRoas(): GetCerByFileTypeId, cerId,fileTypeId,err:", cerId, fileTypeId, err)
						return nil, nil, nil, nil, err
					}
					chainCerAloneTmp := chainmodel.NewChainCerAlone(&chainCerTmp)
					childChainCerAlones = append(childChainCerAlones, *chainCerAloneTmp)
					belogs.Debug("getChildChainCersCrlsMftsRoas(): GetCerByFileTypeId cerId:", cerId,
						"   chainCerAloneTmp.Id:", chainCerAloneTmp.Id, "  len(childChainCerAlones):", len(childChainCerAlones))
				case "crl":
					chainCrl, err := chains.GetCrlByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getChildChainCersCrlsMftsRoas(): GetCrlByFileTypeId, cerId,fileTypeId,err:", cerId, fileTypeId, err)
						return nil, nil, nil, nil, err
					}
					childChainCrls = append(childChainCrls, chainCrl)
				case "mft":
					chainMft, err := chains.GetMftByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getChildChainCersCrlsMftsRoas(): GetMftByFileTypeId, cerId,fileTypeId,err:", cerId, fileTypeId, err)
						return nil, nil, nil, nil, err
					}
					childChainMfts = append(childChainMfts, chainMft)
				case "roa":
					chainRoa, err := chains.GetRoaByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getChildChainCersCrlsMftsRoas(): GetRoaByFileTypeId, cerId,fileTypeId,err:", cerId, fileTypeId, err)
						return nil, nil, nil, nil, err
					}
					childChainRoas = append(childChainRoas, chainRoa)
				}
			}

		}
	}
	belogs.Debug("getChildChainCersCrlsMftsRoas():get all child, cerId:", cerId,
		"  len(childChainCerAlones):", len(childChainCerAlones),
		"  len(childChainCrls):", len(childChainCrls),
		"  len(childChainRoas):", len(childChainRoas),
		"  len(childChainMfts):", len(childChainMfts), "  time(s):", time.Now().Sub(start).Seconds())
	return

}
