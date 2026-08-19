package chainvalidate

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/certutil"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

func getChainRoas(chains *Chains, chainWg *sync.WaitGroup, syncLogId uint64) {

	defer chainWg.Done()
	start := time.Now()

	var roaWg sync.WaitGroup
	chainRoaDatasCh := make(chan []*ChainCertData, conf.Int("chain::getDbConcurrentCount"))
	go callAddRoaToChain(chains, syncLogId, chainRoaDatasCh, &roaWg)
	err := GetChainRoaData(chainRoaDatasCh, &roaWg)
	if err != nil {
		belogs.Error("getChainRoas(): getChainRoaSqlDb fail: ", err)
		close(chainRoaDatasCh)
		return
	}
	belogs.Debug("getChainRoas(): after getChainRoaSqlDb before roaWg.Wait(), time(s):", time.Since(start))
	roaWg.Wait()
	close(chainRoaDatasCh)
	belogs.Info("getChainRoas(): end, syncLogId:", syncLogId, " time(s):", time.Since(start))
}

func callAddRoaToChain(chains *Chains, syncLogId uint64, chainRoaDatasCh chan []*ChainCertData, roaWg *sync.WaitGroup) {
	start := time.Now()
	var index uint64
	for {
		select {
		case chainRoaDatas, ok := <-chainRoaDatasCh:
			belogs.Debug("callAddRoaToChain(): get from chainRoaSqlCh, len(chainRoaSqls):", len(chainRoaDatas),
				"  index:", index, " ok:", ok)
			if ok {
				go addRoaToChain(chains, chainRoaDatas, syncLogId, roaWg)
				index++
				belogs.Debug("callAddRoaToChain(): addRoaToChain index:", index)
			} else {
				belogs.Info("callAddRoaToChain(): close chainRoaSqlsCh, index:", index, "   time(s):", time.Since(start))
				return
			}

		}
	}
}

func addRoaToChain(chains *Chains, chainRoaDatas []*ChainCertData, syncLogId uint64, roaWg *sync.WaitGroup) {
	defer func() {
		roaWg.Done()
		belogs.Debug("len(chainRoaSqls):", len(chainRoaDatas))
	}()

	start := time.Now()
	for i := range chainRoaDatas {
		startOne := time.Now()
		belogs.Debug("addRoaToChain(): chainRoaSql:", jsonutil.MarshalJson(chainRoaDatas[i]), "  syncLogId:", syncLogId)

		chainRoa, err := chainRoaDatas[i].ToChainRoa()
		if err != nil {
			belogs.Error("addRoaToChain(): ToChainRoa fail, chainRoaSql.Id:", chainRoaDatas[i].Id, err)
			return
		}

		chains.AddRoaId(chainRoaDatas[i].Id)
		chains.AddRoa(&chainRoa)
		belogs.Debug("addRoaToChain(): added, chainRoa:", jsonutil.MarshalJson(chainRoa),
			"  syncLogId:", syncLogId, "    chainRoaSql.SyncLogId:", chainRoaDatas[i].SyncLogId,
			"  time(s):", time.Since(startOne))
	}
	belogs.Info("addRoaToChain(): all added, len(chainRoaSqls):", len(chainRoaDatas), "  time(s):", time.Since(start))
	return
}

func validateRoas(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()

	roaIds := chains.RoaIds
	belogs.Debug("validateRoas(): start: len(roaIds):", len(roaIds))

	var roaWg sync.WaitGroup
	chainRoaCh := make(chan int, conf.Int("chain::validateConcurrentCount"))
	for _, roaId := range roaIds {
		roaWg.Add(1)
		chainRoaCh <- 1
		go validateRoa(chains, roaId, &roaWg, chainRoaCh)
	}
	roaWg.Wait()
	close(chainRoaCh)

	belogs.Info("validateRoas(): validate end, len(roaIds):", len(roaIds), "  time(s):", time.Since(start))

}

func validateRoa(chains *Chains, roaId uint64, wg *sync.WaitGroup, chainRoaCh chan int) {
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
		chainRoa.StateModel.JudgeState()
		chains.UpdateFileTypeIdToRoa(&chainRoa)
		belogs.Info("validateRoa(): getRoaParentChainCers fail:", roaId, err)
		return
	}
	belogs.Debug("validateRoa():getRoaParentChainCers, roaId:", roaId, "  len(chainRoa.ParentChainCers):", len(chainRoa.ParentChainCerAlones),
		"  time(s):", time.Since(start))

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
		self := make([]ChainAsn, 0)
		asn := ChainAsn{Asn: chainRoa.Asn}
		self = append(self, asn)
		invalidAsns := asnsIncludeInParents(chainRoa.ParentChainCerAlones, self)
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
	belogs.Debug("validateRoa(): after check ParentChainCerAlones, roaId:", roaId, "  file:", chainRoa.FilePath, chainRoa.FileName,
		"  time(s):", time.Since(start))

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
		belogs.Debug("validateRoa(): stateModel have errors or warnings, roaId :", roaId, "  stateModel:", jsonutil.MarshalJson(chainRoa.StateModel))
	}
	chains.UpdateFileTypeIdToRoa(&chainRoa)
	belogs.Info("validateRoa(): validate one end, UpdateFileTypeIdToRoa, roaId:", roaId,
		" roa file:", chainRoa.FilePath, chainRoa.FileName, "  time(s):", time.Since(start))
}

func getRoaParentChainCers(chains *Chains, roaId uint64) (chainCerAlones []ChainCerAlone, err error) {

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

	chainCerAlones = make([]ChainCerAlone, 0)
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

func getRoaParentChainCer(chains *Chains, roaId uint64) (chainCerAlone ChainCerAlone, err error) {
	start := time.Now()
	chainRoa, err := chains.GetRoaById(roaId)
	if err != nil {
		belogs.Error("getRoaParentChainCer(): GetRoa fail, roaId:", roaId, err)
		return chainCerAlone, err
	}
	belogs.Debug("getRoaParentChainCer(): roaId:", roaId, "  chainRoa.Id:", chainRoa.Id, "  time(s):", time.Since(start))

	//get roa's aki --> parent cer's ski
	if len(chainRoa.Aki) == 0 {
		belogs.Error("getRoaParentChainCer(): chainRoa.Aki is empty fail, roaId:", roaId)
		return chainCerAlone, errors.New("roa's aki is empty")
	}
	aki := chainRoa.Aki
	parentCerSki := aki
	fileTypeId, ok := chains.SkiToFileTypeId[parentCerSki]
	belogs.Debug("getRoaParentChainCer(): roaId:", roaId, "  parentCerSki:", parentCerSki, "  fileTypeId, ok:", fileTypeId, ok)
	if ok {
		parentChainCer, err := chains.GetCerByFileTypeId(fileTypeId)
		belogs.Debug("getRoaParentChainCer(): GetCerByFileTypeId, roaId:", roaId, " fileTypeId:", fileTypeId,
			" parentChainCer.Id:", parentChainCer.Id, "  time(s):", time.Since(start))
		if err != nil {
			belogs.Error("getRoaParentChainCer(): GetCerByFileTypeId fail, roaId:", roaId, " fileTypeId:", fileTypeId, err)
			return chainCerAlone, err
		}
		return *NewChainCerAlone(&parentChainCer), nil

	}
	//  not found parent ,is not error
	belogs.Debug("getRoaParentChainCer(): not found roa's parent cer:", roaId, "  time(s):", time.Since(start))
	return chainCerAlone, nil
}

func updateRoas(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	err := UpdateRoas(chains)
	if err != nil {
		belogs.Error("updateRoas(): UpdateRoas fail:", err)
		return
	}
	belogs.Info("updateRoas(): ok, time(s):", time.Since(start))
}
