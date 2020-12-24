package chainvalidate

import (
	"errors"
	"strings"
	"sync"
	"time"

	belogs "github.com/astaxie/beego/logs"
	certutil "github.com/cpusoft/goutil/certutil"
	conf "github.com/cpusoft/goutil/conf"
	convert "github.com/cpusoft/goutil/convert"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"

	"chainvalidate/db"
	chainmodel "chainvalidate/model"
	"model"
)

func GetChainRoas(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	belogs.Debug("GetChainRoas(): start:")

	chainRoaSqls, err := db.GetChainRoaSqls()
	if err != nil {
		belogs.Error("GetChainRoas(): db.GetChainRoaSqls:", err)
		return
	}
	belogs.Debug("GetChainRoas(): GetChainRoaSqls, len(chainRoaSqls):", len(chainRoaSqls))

	for i := range chainRoaSqls {
		chainRoa := chainRoaSqls[i].ToChainRoa()
		belogs.Debug("GetChainRoas():i, chainRoa:", i, jsonutil.MarshalJson(chainRoa))
		chains.RoaIds = append(chains.RoaIds, chainRoaSqls[i].Id)
		chains.AddRoa(&chainRoa)
	}

	belogs.Debug("GetChainRoas(): end, len(chainRoaSqls):", len(chainRoaSqls), ",   len(chains.RoaIds):", len(chains.RoaIds), "  time(s):", time.Now().Sub(start).Seconds())
	return
}

func ValidateRoas(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()

	roaIds := chains.RoaIds
	belogs.Debug("ValidateRoas(): start: len(roaIds):", len(roaIds))

	var roaWg sync.WaitGroup
	chainRoaCh := make(chan int, conf.Int("chain::chainConcurrentCount"))
	for _, roaId := range roaIds {
		roaWg.Add(1)
		chainRoaCh <- 1
		go validateRoa(chains, roaId, &roaWg, chainRoaCh)
	}
	roaWg.Wait()
	close(chainRoaCh)

	belogs.Info("ValidateRoas(): end, len(roaIds):", len(roaIds), "  time(s):", time.Now().Sub(start).Seconds())

}

func validateRoa(chains *chainmodel.Chains, roaId uint64, wg *sync.WaitGroup, chainRoaCh chan int) {
	defer func() {
		wg.Done()
		<-chainRoaCh
	}()

	start := time.Now()
	chainRoa, err := chains.GetRoaById(roaId)
	if err != nil {
		belogs.Error("validateRoa(): GetRoa fail:", roaId, err)
		return
	}

	chainRoa.ParentChainCerAlones, err = getRoaParentChainCers(chains, roaId)
	if err != nil {
		belogs.Error("validateRoa(): getRoaParentChainCers fail:", roaId, err)
		chainRoa.StateModel.JudgeState()
		chains.UpdateFileTypeIdToRoa(&chainRoa)
		return
	}
	belogs.Debug("validateRoa():chainCer.ParentChainCers, roaId, len(chainRoa.ParentChainCers):", roaId, len(chainRoa.ParentChainCerAlones))

	// if not root cer, should have parent cer
	if len(chainRoa.ParentChainCerAlones) > 0 {
		// get one parent
		parentCer := osutil.JoinPathFile(chainRoa.ParentChainCerAlones[0].FilePath, chainRoa.ParentChainCerAlones[0].FileName)
		roa := osutil.JoinPathFile(chainRoa.FilePath, chainRoa.FileName)
		belogs.Debug("validateRoa(): parentCer:", parentCer, "    roa:", roa)

		// openssl verify roa
		belogs.Debug("validateRoa():before VerifyEeCertByX509,  parentCer:", parentCer,
			"  roa:", roa, "  eeCert:", chainRoa.EeCertStart, chainRoa.EeCertEnd)
		result, err := certutil.VerifyEeCertByX509(parentCer, roa, chainRoa.EeCertStart, chainRoa.EeCertEnd)
		belogs.Debug("validateRoa(): VerifyEeCertByX509 result:", result, err)
		if result != "ok" {
			desc := ""
			if err != nil {
				desc = err.Error()
				belogs.Debug("validateRoa(): VerifyEeCertByX509 fail, roaId:", chainRoa.Id, result, err)
			}
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to be verified by its issuing certificate",
				Detail: desc + "  parent cer file is " + chainRoa.ParentChainCerAlones[0].FileName + ",  roa file is " + chainRoa.FileName}
			// if subject doesnot match ,will just set warning
			if strings.Contains(desc, "issuer name does not match subject from issuing certificate") {
				chainRoa.StateModel.AddWarning(&stateMsg)
			} else {
				chainRoa.StateModel.AddError(&stateMsg)
			}

		}

		// verify ipaddress prefix,if one parent is not found ,found the upper
		// rfc8360: Validation Reconsidered, set error
		invalidIps := IpAddressesIncludeInParents(chainRoa.ParentChainCerAlones, chainRoa.ChainIpAddresses)
		if len(invalidIps) > 0 {
			belogs.Debug("validateRoa(): cer ipaddress is overclaimed, fail, roaId:", chainRoa.Id, jsonutil.MarshalJson(invalidIps))
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "ROA has overclaimed IP address not contained on the issuing certificate",
				Detail: "invalid ip are " + jsonutil.MarshalJson(invalidIps)}
			chainRoa.StateModel.AddError(&stateMsg)
		}

		// verify ipaddress prefix,if one parent is not found ,found the upper
		// rfc8360: Validation Reconsidered, set error
		self := make([]chainmodel.ChainAsn, 0)
		asn := chainmodel.ChainAsn{Asn: chainRoa.Asn}
		self = append(self, asn)
		invalidAsns := AsnsIncludeInParents(chainRoa.ParentChainCerAlones, self)
		if len(invalidAsns) > 0 {
			belogs.Debug("validateRoa(): cer asn is overclaimed, fail, roaId:", chainRoa.Id, jsonutil.MarshalJson(invalidAsns))
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "ROA has overclaimed ASN not contained on the issuing certificate",
				Detail: "invalid asns are " + jsonutil.MarshalJson(invalidAsns)}
			chainRoa.StateModel.AddError(&stateMsg)
		}

	} else {
		belogs.Debug("validateRoa(): roa file has not found parent cer, fail, chainRoa.Id, roaId:", chainRoa.Id, roaId)
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Its issuing certificate no longer exists",
			Detail: ""}
		chainRoa.StateModel.AddError(&stateMsg)

	}

	if len(chainRoa.ChainSnInCrlRevoked.CrlFileName) > 0 {
		belogs.Debug("validateRoa(): roa ee file is founded in crl's revoked cer list:",
			chainRoa.Id, jsonutil.MarshalJson(chainRoa.ChainSnInCrlRevoked.CrlFileName))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail: "The EE of this ROA is found on the revocation list of CRL",
			Detail: chainRoa.FileName + " is in " + chainRoa.ChainSnInCrlRevoked.CrlFileName + " revoked cer list, " +
				" and revoked time is " + convert.Time2StringZone(chainRoa.ChainSnInCrlRevoked.RevocationTime)}
		chainRoa.StateModel.AddError(&stateMsg)
	}

	chainRoa.StateModel.JudgeState()
	belogs.Debug("validateRoa(): roaId, stateModel:", roaId, chainRoa.StateModel)
	if chainRoa.StateModel.State != "valid" {
		belogs.Info("validateRoa(): stateModel have errors or warnings, roaId :", roaId, "  stateModel:", jsonutil.MarshalJson(chainRoa.StateModel))
	}
	chains.UpdateFileTypeIdToRoa(&chainRoa)
	belogs.Debug("validateRoa():end UpdateFileTypeIdToRoa, roaId:", roaId, "  time(s):", time.Now().Sub(start).Seconds())
}

func getRoaParentChainCers(chains *chainmodel.Chains, roaId uint64) (chainCerAlones []chainmodel.ChainCerAlone, err error) {

	parentChainCerAlone, err := getRoaParentChainCer(chains, roaId)
	if err != nil {
		belogs.Error("getRoaParentChainCers(): getRoaParentChainCer, roaId:", roaId, err)
		return nil, err
	}
	belogs.Debug("getRoaParentChainCers(): roaId:", roaId, "  parentChainCerAlone.Id:", parentChainCerAlone.Id)

	if parentChainCerAlone.Id == 0 {
		belogs.Debug("getRoaParentChainCers(): parentChainCer is not found , roaId :", roaId)
		return chainCerAlones, nil
	}

	chainCerAlones = make([]chainmodel.ChainCerAlone, 0)
	chainCerAlones = append(chainCerAlones, parentChainCerAlone)
	chainCerAlonesTmp, err := GetCerParentChainCers(chains, parentChainCerAlone.Id)
	if err != nil {
		belogs.Error("getRoaParentChainCers(): GetCerParentChainCers, roaId:", roaId, "   parentChainCerAlone.Id:", parentChainCerAlone.Id, err)
		return nil, err
	}
	chainCerAlones = append(chainCerAlones, chainCerAlonesTmp...)
	belogs.Debug("getRoaParentChainCers():roaId, len(chainCerAlones):", roaId, len(chainCerAlones))
	return chainCerAlones, nil
}
func getRoaParentChainCer(chains *chainmodel.Chains, roaId uint64) (chainCerAlone chainmodel.ChainCerAlone, err error) {
	chainRoa, err := chains.GetRoaById(roaId)
	if err != nil {
		belogs.Error("getRoaParentChainCer(): GetRoa, roaId:", roaId, err)
		return chainCerAlone, err
	}
	belogs.Debug("getRoaParentChainCer(): roaId:", roaId, "  chainRoa.Id:", chainRoa.Id)

	//get roa's aki --> parent cer's ski
	if len(chainRoa.Aki) == 0 {
		belogs.Error("getRoaParentChainCer(): chainRoa.Aki is empty, fail:", roaId)
		return chainCerAlone, errors.New("roa's aki is empty")
	}
	aki := chainRoa.Aki
	parentCerSki := aki
	fileTypeId, ok := chains.SkiToFileTypeId[parentCerSki]
	belogs.Debug("getRoaParentChainCer(): roaId:", roaId, "  parentCerSki:", parentCerSki, "  fileTypeId, ok:", fileTypeId, ok)
	if ok {
		parentChainCer, err := chains.GetCerByFileTypeId(fileTypeId)
		belogs.Debug("getRoaParentChainCer(): GetCerByFileTypeId, roaId, fileTypeId, parentChainCer.Id:", roaId, fileTypeId, parentChainCer.Id)
		if err != nil {
			belogs.Error("getRoaParentChainCer(): GetCerByFileTypeId, roaId,fileTypeId, fail:", roaId, fileTypeId, err)
			return chainCerAlone, err
		}
		return *chainmodel.NewChainCerAlone(&parentChainCer), nil

	}
	//  not found parent ,is not error
	belogs.Debug("getRoaParentChainCer(): not found roa's parent cer:", roaId)
	return chainCerAlone, nil
}
