package chainvalidate

import (
	"errors"
	"strings"
	"sync"
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/certutil"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

func getChainAsas(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	belogs.Debug("getChainAsas(): start:")

	chainAsaSqls, err := getChainAsaSqlsDb()
	if err != nil {
		belogs.Error("getChainAsas(): getChainAsaSqlsDb:", err)
		return
	}
	belogs.Debug("getChainAsas(): getChainAsaSqlsDb, len(chainAsaSqls):", len(chainAsaSqls))

	for i := range chainAsaSqls {
		chainAsa := chainAsaSqls[i].ToChainAsa()
		belogs.Debug("getChainAsas():i, chainAsa:", i, jsonutil.MarshalJson(chainAsa))
		chains.AsaIds = append(chains.AsaIds, chainAsaSqls[i].Id)
		chains.AddAsa(&chainAsa)
	}

	belogs.Debug("getChainAsas(): end, len(chainAsaSqls):", len(chainAsaSqls), ",   len(chains.AsaIds):", len(chains.AsaIds), "  time(s):", time.Since(start))
	return
}

func validateAsas(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()

	asaIds := chains.AsaIds
	belogs.Debug("validateAsas(): start: len(asaIds):", len(asaIds))

	var asaWg sync.WaitGroup
	chainAsaCh := make(chan int, conf.Int("chain::chainConcurrentCount"))
	for _, asaId := range asaIds {
		asaWg.Add(1)
		chainAsaCh <- 1
		go validateAsa(chains, asaId, &asaWg, chainAsaCh)
	}
	asaWg.Wait()
	close(chainAsaCh)

	belogs.Info("validateAsas(): end, len(asaIds):", len(asaIds), "  time(s):", time.Since(start))

}

func validateAsa(chains *Chains, asaId uint64, wg *sync.WaitGroup, chainAsaCh chan int) {
	defer func() {
		wg.Done()
		<-chainAsaCh
	}()

	start := time.Now()
	chainAsa, err := chains.GetAsaById(asaId)
	if err != nil {
		belogs.Error("validateAsa(): GetAsa fail:", asaId, err)
		return
	}

	chainAsa.ParentChainCerAlones, err = getAsaParentChainCers(chains, asaId)
	if err != nil {
		belogs.Info("validateAsa(): getAsaParentChainCers fail:", asaId, err)
		chainAsa.StateModel.JudgeState()
		chains.UpdateFileTypeIdToAsa(&chainAsa)
		return
	}
	belogs.Debug("validateAsa():chainCer.ParentChainCers, asaId, len(chainAsa.ParentChainCers):", asaId, len(chainAsa.ParentChainCerAlones))

	// if not root cer, should have parent cer
	if len(chainAsa.ParentChainCerAlones) > 0 {
		// get one parent
		parentCer := osutil.JoinPathFile(chainAsa.ParentChainCerAlones[0].FilePath, chainAsa.ParentChainCerAlones[0].FileName)
		asa := osutil.JoinPathFile(chainAsa.FilePath, chainAsa.FileName)
		belogs.Debug("validateAsa(): parentCer:", parentCer, "    asa:", asa)

		// openssl verify asa
		belogs.Debug("validateAsa():before VerifyEeCertByX509,  parentCer:", parentCer,
			"  asa:", asa, "  eeCert:", chainAsa.EeCertStart, chainAsa.EeCertEnd)
		result, err := certutil.VerifyEeCertByX509(parentCer, asa, chainAsa.EeCertStart, chainAsa.EeCertEnd)
		belogs.Debug("validateAsa(): VerifyEeCertByX509 result:", result, err)
		if result != "ok" {
			desc := ""
			if err != nil {
				desc = err.Error()
				belogs.Debug("validateAsa(): VerifyEeCertByX509 fail, asaId:", chainAsa.Id, result, err)
			}
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to be verified by its issuing certificate",
				Detail: desc + "  parent cer file is " + chainAsa.ParentChainCerAlones[0].FileName + ",  asa file is " + chainAsa.FileName}
			// if subject doesnot match ,will just set warning
			if strings.Contains(desc, "issuer name does not match subject from issuing certificate") {
				chainAsa.StateModel.AddWarning(&stateMsg)
			} else {
				chainAsa.StateModel.AddError(&stateMsg)
			}

		}

	} else {
		belogs.Debug("validateAsa(): asa file has not found parent cer, fail, chainAsa.Id, asaId:", chainAsa.Id, asaId)
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Its issuing certificate no longer exists",
			Detail: ""}
		chainAsa.StateModel.AddError(&stateMsg)

	}

	if len(chainAsa.ChainSnInCrlRevoked.CrlFileName) > 0 {
		belogs.Debug("validateAsa(): asa ee file is founded in crl's revoked cer list:",
			chainAsa.Id, jsonutil.MarshalJson(chainAsa.ChainSnInCrlRevoked.CrlFileName))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail: "The EE of this ROA is found on the revocation list of CRL",
			Detail: chainAsa.FileName + " is in " + chainAsa.ChainSnInCrlRevoked.CrlFileName + " revoked cer list, " +
				" and revoked time is " + convert.Time2StringZone(chainAsa.ChainSnInCrlRevoked.RevocationTime)}
		chainAsa.StateModel.AddError(&stateMsg)
	}

	chainAsa.StateModel.JudgeState()
	belogs.Debug("validateAsa(): asaId, stateModel:", asaId, chainAsa.StateModel)
	if chainAsa.StateModel.State != "valid" {
		belogs.Debug("validateAsa(): stateModel have errors or warnings, asaId :", asaId, "  stateModel:", jsonutil.MarshalJson(chainAsa.StateModel))
	}
	chains.UpdateFileTypeIdToAsa(&chainAsa)
	belogs.Debug("validateAsa():end UpdateFileTypeIdToAsa, asaId:", asaId, "  time(s):", time.Since(start))
}

func getAsaParentChainCers(chains *Chains, asaId uint64) (chainCerAlones []ChainCerAlone, err error) {

	parentChainCerAlone, err := getAsaParentChainCer(chains, asaId)
	if err != nil {
		belogs.Error("getAsaParentChainCers(): getAsaParentChainCer, asaId:", asaId, err)
		return nil, err
	}
	belogs.Debug("getAsaParentChainCers(): asaId:", asaId, "  parentChainCerAlone.Id:", parentChainCerAlone.Id)

	if parentChainCerAlone.Id == 0 {
		belogs.Debug("getAsaParentChainCers(): parentChainCer is not found , asaId :", asaId)
		return chainCerAlones, nil
	}

	chainCerAlones = make([]ChainCerAlone, 0)
	chainCerAlones = append(chainCerAlones, parentChainCerAlone)
	chainCerAlonesTmp, err := GetCerParentChainCers(chains, parentChainCerAlone.Id)
	if err != nil {
		belogs.Error("getAsaParentChainCers(): GetCerParentChainCers, asaId:", asaId, "   parentChainCerAlone.Id:", parentChainCerAlone.Id, err)
		return nil, err
	}
	chainCerAlones = append(chainCerAlones, chainCerAlonesTmp...)
	belogs.Debug("getAsaParentChainCers():asaId, len(chainCerAlones):", asaId, len(chainCerAlones))
	return chainCerAlones, nil
}
func getAsaParentChainCer(chains *Chains, asaId uint64) (chainCerAlone ChainCerAlone, err error) {
	chainAsa, err := chains.GetAsaById(asaId)
	if err != nil {
		belogs.Error("getAsaParentChainCer(): GetAsa, asaId:", asaId, err)
		return chainCerAlone, err
	}
	belogs.Debug("getAsaParentChainCer(): asaId:", asaId, "  chainAsa.Id:", chainAsa.Id)

	//get asa's aki --> parent cer's ski
	if len(chainAsa.Aki) == 0 {
		belogs.Error("getAsaParentChainCer(): chainAsa.Aki is empty, fail:", asaId)
		return chainCerAlone, errors.New("asa's aki is empty")
	}
	aki := chainAsa.Aki
	parentCerSki := aki
	fileTypeId, ok := chains.SkiToFileTypeId[parentCerSki]
	belogs.Debug("getAsaParentChainCer(): asaId:", asaId, "  parentCerSki:", parentCerSki, "  fileTypeId, ok:", fileTypeId, ok)
	if ok {
		parentChainCer, err := chains.GetCerByFileTypeId(fileTypeId)
		belogs.Debug("getAsaParentChainCer(): GetCerByFileTypeId, asaId, fileTypeId, parentChainCer.Id:", asaId, fileTypeId, parentChainCer.Id)
		if err != nil {
			belogs.Error("getAsaParentChainCer(): GetCerByFileTypeId, asaId,fileTypeId, fail:", asaId, fileTypeId, err)
			return chainCerAlone, err
		}
		return *NewChainCerAlone(&parentChainCer), nil

	}
	//  not found parent ,is not error
	belogs.Debug("getAsaParentChainCer(): not found asa's parent cer:", asaId)
	return chainCerAlone, nil
}
