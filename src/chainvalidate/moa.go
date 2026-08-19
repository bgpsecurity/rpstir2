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

func getChainMoas(chains *Chains, chainWg *sync.WaitGroup, syncLogId uint64) {

	defer chainWg.Done()
	start := time.Now()

	var moaWg sync.WaitGroup
	chainRoaDatasCh := make(chan []*ChainCertData, conf.Int("chain::getDbConcurrentCount"))
	go callAddMoaToChain(chains, syncLogId, chainRoaDatasCh, &moaWg)
	err := GetChainMoaData(chainRoaDatasCh, &moaWg)
	if err != nil {
		belogs.Error("getChainMoas(): getChainRoaSqlDb fail: ", err)
		close(chainRoaDatasCh)
		return
	}
	belogs.Debug("getChainMoas(): after getChainRoaSqlDb before moaWg.Wait(), time(s):", time.Since(start))
	moaWg.Wait()
	close(chainRoaDatasCh)
	belogs.Info("getChainMoas(): end, syncLogId:", syncLogId, " time(s):", time.Since(start))
}

func callAddMoaToChain(chains *Chains, syncLogId uint64, chainMoaDatasCh chan []*ChainCertData, roaWg *sync.WaitGroup) {
	start := time.Now()
	var index uint64
	for {
		select {
		case chainMoaDatas, ok := <-chainMoaDatasCh:
			belogs.Debug("callAddMoaToChain(): get from chainMoaSqlCh, len(chainMoaSqls):", len(chainMoaDatas),
				"  index:", index, " ok:", ok)
			if ok {
				go addMoaToChain(chains, chainMoaDatas, syncLogId, roaWg)
				index++
				belogs.Debug("callAddMoaToChain(): addMoaToChain index:", index)
			} else {
				belogs.Info("callAddMoaToChain(): close chainMoaSqlsCh, index:", index, "   time(s):", time.Since(start))
				return
			}

		}
	}
}

//func callAddMoaToChain(chains *Chains, syncLogId uint64, chainMoaSqlsCh chan []*ChainCertSql, moaWg *sync.WaitGroup) {
//	start := time.Now()
//	var index uint64
//	for {
//		select {
//		case chainMoaSqls, ok := <-chainMoaSqlsCh:
//			belogs.Debug("callAddMoaToChain(): get from chainMoaSqlsCh, len(chainMoaSqls):", len(chainMoaSqls),
//				"  index:", index, "  ok:", ok)
//			if ok {
//				go addMoaToChain(chains, chainMoaSqls, syncLogId, moaWg)
//				index++
//				belogs.Debug("callAddMoaToChain(): addMoaToChain index:", index)
//			} else {
//				belogs.Info("callAddMoaToChain(): close chainMoaSqlsCh, index:", index, "   time(s):", time.Since(start))
//				return
//			}
//		}
//	}
//}

func addMoaToChain(chains *Chains, chainMoaDatas []*ChainCertData, syncLogId uint64, moaWg *sync.WaitGroup) {
	defer func() {
		moaWg.Done()
		belogs.Debug("addMoaToChain(): moaWg.Done(), len(chainMoaSqls):", len(chainMoaDatas))
	}()
	start := time.Now()

	for i := range chainMoaDatas {
		startOne := time.Now()
		belogs.Debug("addMoaToChain(): chainMoaSql:", jsonutil.MarshalJson(chainMoaDatas[i]), "  syncLogId:", syncLogId)

		chainMoa, err := chainMoaDatas[i].ToChainMoa()
		if err != nil {
			belogs.Error("addMoaToChain(): ToChainMoa fail, chainMoaSql.Id:", chainMoaDatas[i].Id, err)
			return
		}

		chains.AddMoaId(chainMoaDatas[i].Id)
		chains.AddMoa(&chainMoa)
		belogs.Debug("addMoaToChain(): added, chainMoa:", jsonutil.MarshalJson(chainMoa),
			"  syncLogId:", syncLogId, "    chainMoaSql.SyncLogId:", chainMoaDatas[i].SyncLogId,
			"  time(s):", time.Since(startOne))
	}
	belogs.Info("addMoaToChain(): all added, len(chainMoaDatas):", len(chainMoaDatas), "  time(s):", time.Since(start))
	return
}

func validateMoas(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()

	moaIds := chains.MoaIds
	belogs.Debug("validateMoas(): start: len(moaIds):", len(moaIds))

	var moaWg sync.WaitGroup
	chainMoaCh := make(chan int, conf.Int("chain::validateConcurrentCount"))
	for _, moaId := range moaIds {
		moaWg.Add(1)
		chainMoaCh <- 1
		go validateMoa(chains, moaId, &moaWg, chainMoaCh)
	}
	moaWg.Wait()
	close(chainMoaCh)

	belogs.Info("validateMoas(): validate end, len(moaIds):", len(moaIds), "  time(s):", time.Since(start))

}

func validateMoa(chains *Chains, moaId uint64, wg *sync.WaitGroup, chainMoaCh chan int) {
	defer func() {
		wg.Done()
		<-chainMoaCh
	}()

	start := time.Now()
	chainMoa, err := chains.GetMoaById(moaId)
	if err != nil {
		belogs.Error("validateMoa(): GetMoa fail:", moaId, err)
		return
	}

	chainMoa.ParentChainCerAlones, err = getMoaParentChainCers(chains, moaId)
	if err != nil {
		belogs.Info("validateMoa(): getMoaParentChainCers fail:", moaId, err)
		chainMoa.StateModel.JudgeState()
		chains.UpdateFileTypeIdToMoa(&chainMoa)
		return
	}
	belogs.Debug("validateMoa():getMoaParentChainCers, moaId:", moaId, "   len(chainMoa.ParentChainCers):", len(chainMoa.ParentChainCerAlones))

	// if not root cer, should have parent cer
	if len(chainMoa.ParentChainCerAlones) > 0 {
		// get one parent
		parentCer := osutil.JoinPathFile(chainMoa.ParentChainCerAlones[0].FilePath, chainMoa.ParentChainCerAlones[0].FileName)
		moa := osutil.JoinPathFile(chainMoa.FilePath, chainMoa.FileName)
		belogs.Debug("validateMoa(): parentCer:", parentCer, "    moa:", moa)

		// openssl verify moa
		belogs.Debug("validateMoa():before VerifyEeCertByX509,  parentCer:", parentCer,
			"  moa:", moa, "  eeCert:", chainMoa.EeCertStart, chainMoa.EeCertEnd)
		result, err := certutil.VerifyEeCertByX509(parentCer, moa, chainMoa.EeCertStart, chainMoa.EeCertEnd)
		belogs.Debug("validateMoa(): VerifyEeCertByX509 result:", result, err)
		if result != "ok" {
			desc := ""
			if err != nil {
				desc = err.Error()
				belogs.Debug("validateMoa(): VerifyEeCertByX509 fail, moaId:", chainMoa.Id, result, err)
			}
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to be verified by its issuing certificate",
				Detail: desc + "  parent cer file is " + chainMoa.ParentChainCerAlones[0].FileName + ",  moa file is " + chainMoa.FileName}
			// if subject doesnot match ,will just set warning
			if strings.Contains(desc, "issuer name does not match subject from issuing certificate") {
				chainMoa.StateModel.AddWarning(&stateMsg)
			} else {
				chainMoa.StateModel.AddError(&stateMsg)
			}

		}

	} else {
		belogs.Debug("validateMoa(): moa file has not found parent cer, fail, chainMoa.Id, moaId:", chainMoa.Id, moaId)
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Its issuing certificate no longer exists",
			Detail: ""}
		chainMoa.StateModel.AddError(&stateMsg)

	}

	if len(chainMoa.ChainSnInCrlRevoked.CrlFileName) > 0 {
		belogs.Debug("validateMoa(): moa ee file is founded in crl's revoked cer list:",
			chainMoa.Id, jsonutil.MarshalJson(chainMoa.ChainSnInCrlRevoked.CrlFileName))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail: "The EE of this ROA is found on the revocation list of CRL",
			Detail: chainMoa.FileName + " is in " + chainMoa.ChainSnInCrlRevoked.CrlFileName + " revoked cer list, " +
				" and revoked time is " + convert.Time2StringZone(chainMoa.ChainSnInCrlRevoked.RevocationTime)}
		chainMoa.StateModel.AddError(&stateMsg)
	}

	chainMoa.StateModel.JudgeState()
	belogs.Debug("validateMoa(): moaId, stateModel:", moaId, chainMoa.StateModel)
	if chainMoa.StateModel.State != "valid" {
		belogs.Debug("validateMoa(): stateModel have errors or warnings, moaId :", moaId, "  stateModel:", jsonutil.MarshalJson(chainMoa.StateModel))
	}
	chains.UpdateFileTypeIdToMoa(&chainMoa)
	belogs.Info("validateMoa(): validate one end, UpdateFileTypeIdToMoa, moaId:", moaId,
		" moa file:", chainMoa.FilePath, chainMoa.FileName, "  time(s):", time.Since(start))
}

func basicValidateMoa(chains *Chains, moaId uint64, wg *sync.WaitGroup, chainMoaCh chan int) {
	var chainMoa ChainMoa

	defer func() {

		if chainMoa.StateModel.State != "valid" {
			belogs.Debug("validateMoa(): stateModel have errors or warnings, moaId :", moaId, "  stateModel:", jsonutil.MarshalJson(chainMoa.StateModel))

		}

		wg.Done()
		<-chainMoaCh
	}()

	start := time.Now()
	chainMoa, err := chains.GetMoaById(moaId)
	if err != nil {
		belogs.Error("validateMoa(): GetMoa fail:", moaId, err)
		return
	}

	chainMoa.ParentChainCerAlones, err = getMoaParentChainCers(chains, moaId)
	if err != nil {
		belogs.Info("validateMoa(): getMoaParentChainCers fail:", moaId, err)
		chainMoa.StateModel.JudgeState()
		chains.UpdateFileTypeIdToMoa(&chainMoa)
		return
	}
	belogs.Debug("validateMoa():getMoaParentChainCers, moaId:", moaId, "   len(chainMoa.ParentChainCers):", len(chainMoa.ParentChainCerAlones))

	// if not root cer, should have parent cer
	if len(chainMoa.ParentChainCerAlones) > 0 {
		// get one parent
		parentCer := osutil.JoinPathFile(chainMoa.ParentChainCerAlones[0].FilePath, chainMoa.ParentChainCerAlones[0].FileName)
		moa := osutil.JoinPathFile(chainMoa.FilePath, chainMoa.FileName)
		belogs.Debug("validateMoa(): parentCer:", parentCer, "    moa:", moa)

		// openssl verify moa
		belogs.Debug("validateMoa():before VerifyEeCertByX509,  parentCer:", parentCer,
			"  moa:", moa, "  eeCert:", chainMoa.EeCertStart, chainMoa.EeCertEnd)
		result, err := certutil.VerifyEeCertByX509(parentCer, moa, chainMoa.EeCertStart, chainMoa.EeCertEnd)
		belogs.Debug("validateMoa(): VerifyEeCertByX509 result:", result, err)
		if result != "ok" {
			desc := ""
			if err != nil {
				desc = err.Error()
				belogs.Debug("validateMoa(): VerifyEeCertByX509 fail, moaId:", chainMoa.Id, result, err)
			}
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to be verified by its issuing certificate",
				Detail: desc + "  parent cer file is " + chainMoa.ParentChainCerAlones[0].FileName + ",  moa file is " + chainMoa.FileName}
			// if subject doesnot match ,will just set warning
			if strings.Contains(desc, "issuer name does not match subject from issuing certificate") {
				chainMoa.StateModel.AddWarning(&stateMsg)
			} else {
				chainMoa.StateModel.AddError(&stateMsg)
			}

		}

	} else {
		belogs.Debug("validateMoa(): moa file has not found parent cer, fail, chainMoa.Id, moaId:", chainMoa.Id, moaId)
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Its issuing certificate no longer exists",
			Detail: ""}
		chainMoa.StateModel.AddError(&stateMsg)

	}

	if len(chainMoa.ChainSnInCrlRevoked.CrlFileName) > 0 {
		belogs.Debug("validateMoa(): moa ee file is founded in crl's revoked cer list:",
			chainMoa.Id, jsonutil.MarshalJson(chainMoa.ChainSnInCrlRevoked.CrlFileName))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail: "The EE of this ROA is found on the revocation list of CRL",
			Detail: chainMoa.FileName + " is in " + chainMoa.ChainSnInCrlRevoked.CrlFileName + " revoked cer list, " +
				" and revoked time is " + convert.Time2StringZone(chainMoa.ChainSnInCrlRevoked.RevocationTime)}
		chainMoa.StateModel.AddError(&stateMsg)
	}

	chainMoa.StateModel.JudgeState()
	belogs.Debug("validateMoa(): moaId, stateModel:", moaId, chainMoa.StateModel)
	//if chainMoa.StateModel.State != "valid" {
	//	belogs.Debug("validateMoa(): stateModel have errors or warnings, moaId :", moaId, "  stateModel:", jsonutil.MarshalJson(chainMoa.StateModel))
	//}
	chains.UpdateFileTypeIdToMoa(&chainMoa)
	belogs.Info("validateMoa(): validate one end, UpdateFileTypeIdToMoa, moaId:", moaId,
		" moa file:", chainMoa.FilePath, chainMoa.FileName, "  time(s):", time.Since(start))

	return
}

func getMoaParentChainCers(chains *Chains, moaId uint64) (chainCerAlones []ChainCerAlone, err error) {

	parentChainCerAlone, err := getMoaParentChainCer(chains, moaId)
	if err != nil {
		belogs.Error("getMoaParentChainCers(): getMoaParentChainCer, moaId:", moaId, err)
		return nil, err
	}
	belogs.Debug("getMoaParentChainCers(): moaId:", moaId, "  parentChainCerAlone.Id:", parentChainCerAlone.Id)

	if parentChainCerAlone.Id == 0 {
		belogs.Debug("getMoaParentChainCers(): parentChainCer is not found , moaId :", moaId)
		return chainCerAlones, nil
	}

	chainCerAlones = make([]ChainCerAlone, 0)
	chainCerAlones = append(chainCerAlones, parentChainCerAlone)
	chainCerAlonesTmp, err := GetCerParentChainCers(chains, parentChainCerAlone.Id)
	if err != nil {
		belogs.Error("getMoaParentChainCers(): GetCerParentChainCers, moaId:", moaId, "   parentChainCerAlone.Id:", parentChainCerAlone.Id, err)
		return nil, err
	}
	chainCerAlones = append(chainCerAlones, chainCerAlonesTmp...)
	belogs.Debug("getMoaParentChainCers():moaId, len(chainCerAlones):", moaId, len(chainCerAlones))
	return chainCerAlones, nil
}
func getMoaParentChainCer(chains *Chains, moaId uint64) (chainCerAlone ChainCerAlone, err error) {
	chainMoa, err := chains.GetMoaById(moaId)
	if err != nil {
		belogs.Error("getMoaParentChainCer(): GetMoa, moaId:", moaId, err)
		return chainCerAlone, err
	}
	belogs.Debug("getMoaParentChainCer(): moaId:", moaId, "  chainMoa.Id:", chainMoa.Id)

	//get moa's aki --> parent cer's ski
	if len(chainMoa.Aki) == 0 {
		belogs.Error("getMoaParentChainCer(): chainMoa.Aki is empty, fail:", moaId)
		return chainCerAlone, errors.New("moa's aki is empty")
	}
	aki := chainMoa.Aki
	parentCerSki := aki
	fileTypeId, ok := chains.SkiToFileTypeId[parentCerSki]
	belogs.Debug("getMoaParentChainCer(): moaId:", moaId, "  parentCerSki:", parentCerSki, "  fileTypeId, ok:", fileTypeId, ok)
	if ok {
		parentChainCer, err := chains.GetCerByFileTypeId(fileTypeId)
		belogs.Debug("getMoaParentChainCer(): GetCerByFileTypeId, moaId, fileTypeId, parentChainCer.Id:", moaId, fileTypeId, parentChainCer.Id)
		if err != nil {
			belogs.Error("getMoaParentChainCer(): GetCerByFileTypeId, moaId,fileTypeId, fail:", moaId, fileTypeId, err)
			return chainCerAlone, err
		}
		return *NewChainCerAlone(&parentChainCer), nil

	}
	//  not found parent ,is not error
	belogs.Debug("getMoaParentChainCer(): not found moa's parent cer:", moaId)
	return chainCerAlone, nil
}

func updateMoas(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	err := UpdateMoas(chains)
	if err != nil {
		belogs.Error("updateMoas(): UpdateMoas fail:", err)
		return
	}
	belogs.Info("updateMoas(): ok,  time(s):", time.Since(start))
}
