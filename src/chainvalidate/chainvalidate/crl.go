package chainvalidate

import (
	"errors"
	"strings"
	"sync"
	"time"

	belogs "github.com/astaxie/beego/logs"
	certutil "github.com/cpusoft/goutil/certutil"
	conf "github.com/cpusoft/goutil/conf"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"

	"chainvalidate/db"
	chainmodel "chainvalidate/model"
	"model"
)

func GetChainCrls(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	belogs.Debug("GetChainCrls(): start:")

	chainCrlSqls, err := db.GetChainCrlSqls()
	if err != nil {
		belogs.Error("GetChainCrls(): db.GetChainCrlSqls:", err)
		return
	}
	belogs.Debug("GetChainCrls(): GetChainCers, len(chainCrlSqls):", len(chainCrlSqls))

	for i := range chainCrlSqls {
		chainCrl := chainCrlSqls[i].ToChainCrl()
		belogs.Debug("GetChainCrls():i, chainCrl:", i, jsonutil.MarshalJson(chainCrl))
		chains.CrlIds = append(chains.CrlIds, chainCrlSqls[i].Id)
		chains.AddCrl(&chainCrl)
	}

	belogs.Debug("GetChainCrls(): end, len(chainCrlSqls):", len(chainCrlSqls), ",   len(chains.CrlIds):", len(chains.CrlIds), "  time(s):", time.Now().Sub(start).Seconds())
	return
}

func ValidateCrls(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()

	crlIds := chains.CrlIds
	belogs.Debug("ValidateCrls(): start: len(crlIds):", len(crlIds))

	var crlWg sync.WaitGroup
	chainCrlCh := make(chan int, conf.Int("chain::chainConcurrentCount"))
	for _, crlId := range crlIds {
		crlWg.Add(1)
		chainCrlCh <- 1
		go validateCrl(chains, crlId, &crlWg, chainCrlCh)
	}
	crlWg.Wait()
	close(chainCrlCh)

	belogs.Info("ValidateCrls(): end, len(crlIds):", len(crlIds), "  time(s):", time.Now().Sub(start).Seconds())
}

func validateCrl(chains *chainmodel.Chains, crlId uint64, wg *sync.WaitGroup, chainCrlCh chan int) {
	defer func() {
		wg.Done()
		<-chainCrlCh
	}()

	start := time.Now()
	chainCrl, err := chains.GetCrlById(crlId)
	if err != nil {
		belogs.Error("validateCrl(): crlId fail:", crlId, err)
		return
	}
	// set parent cer
	chainCrl.ParentChainCerAlones, err = getCrlParentChainCers(chains, crlId)
	if err != nil {
		belogs.Error("getChainCrl(): GetCrlParentChainCer fail:", crlId, err)
		chainCrl.StateModel.JudgeState()
		chains.UpdateFileTypeIdToCrl(&chainCrl)
		return
	}
	belogs.Debug("validateCrl():chainCrl.ParentChainCer, crlId,  len(chainCrl.ParentChainCerAlones):", crlId, len(chainCrl.ParentChainCerAlones))

	// exist parent cer
	if len(chainCrl.ParentChainCerAlones) > 0 {
		// get one parent
		parentCer := osutil.JoinPathFile(chainCrl.ParentChainCerAlones[0].FilePath, chainCrl.ParentChainCerAlones[0].FileName)
		crl := osutil.JoinPathFile(chainCrl.FilePath, chainCrl.FileName)
		belogs.Debug("validateCrl(): parentCer:", parentCer, "    crl:", crl)

		// openssl verify crl
		result, err := certutil.VerifyCrlByX509(parentCer, crl)
		belogs.Debug("validateCrl(): VerifyCrlByX509 result:", result, err)
		if result != "ok" {
			desc := ""
			if err != nil {
				desc = err.Error()
				belogs.Debug("validateCrl(): verify crl by parent cer fail, fail, crlId:",
					"  crl:", crl, "   parentCer:", parentCer, "  chainCrl.Id:", chainCrl.Id, err)
			}
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to be verified by its issuing certificate",
				Detail: desc + "  parent cer file is " + chainCrl.ParentChainCerAlones[0].FileName + ",  crl file is " + chainCrl.FileName}
			// if subject doesnot match ,will just set warning
			if strings.Contains(desc, "issuer name does not match subject from issuing certificate") {
				chainCrl.StateModel.AddWarning(&stateMsg)
			} else {
				chainCrl.StateModel.AddError(&stateMsg)
			}
		}

	} else {
		belogs.Debug("validateCrl(): crl file has not found parent cer, fail, chainCrl.Id, crlId:", chainCrl.Id, crlId)
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Its issuing certificate no longer exists",
			Detail: ""}
		chainCrl.StateModel.AddError(&stateMsg)

	}

	// cer in crl(by sn) should not exists
	if len(chainCrl.ShouldRevokedCerts) > 0 {
		belogs.Debug("validateCrl(): len(chainCrl.ShouldRevokedCerts) > 0, crlId:", chainCrl.Id, jsonutil.MarshalJson(chainCrl.ShouldRevokedCerts))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Files on revocation list of CRL still exists",
			Detail: "crl file is " + chainCrl.FileName + ", and should be revoked cers/roas/mfts are " + strings.Join(chainCrl.ShouldRevokedCerts, ", ")}
		chainCrl.StateModel.AddError(&stateMsg)
	}

	// check same aki crl files, compare crl number
	sameAkiChainCrls, err := getSameAkiChainCrls(chains, crlId)
	if err != nil {
		belogs.Error("validateCrl():GetSameAkiCrlFiles fail, aki:", chainCrl.Aki)
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Fail to get CRL under specific AKI",
			Detail: err.Error()}
		chainCrl.StateModel.AddError(&stateMsg)
	} else {

		belogs.Debug("validateCrl():GetSameAkiCrlFiles aki:", chainCrl.Aki,
			" self is ", chainCrl.FileName,
			" sameAkiChainCrls:", jsonutil.MarshalJson(sameAkiChainCrls))
		if len(sameAkiChainCrls) == 1 {
			// filename shoud be equal
			if sameAkiChainCrls[0].FileName != chainCrl.FileName {
				belogs.Error("validateCrl():same crl files is not self, aki:", sameAkiChainCrls[0].FileName, chainCrl.FileName, chainCrl.Aki)
				stateMsg := model.StateMsg{Stage: "chainvalidate",
					Fail:   "Fail to get CRL under specific AKI",
					Detail: "aki is" + chainCrl.Aki + "  fileName is " + chainCrl.FileName}
				chainCrl.StateModel.AddError(&stateMsg)
			}
		} else if len(sameAkiChainCrls) == 0 {
			belogs.Debug("validateCrl():same mft files is zero, aki:", chainCrl.Aki)
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to get CRL under specific AKI",
				Detail: "aki is " + chainCrl.Aki + ",  fileName should be " + chainCrl.FileName}
			chainCrl.StateModel.AddError(&stateMsg)
		} else {
			belogs.Debug("validateCrl():more than one same aki crl files, ",
				chainCrl.Aki, chainCrl.FileName, chainCrl.CrlNumber, "  sameAkiChainCrls: ", jsonutil.MarshalJson(sameAkiChainCrls))
			// smaller/older are more ahead
			smallerFiles := make([]chainmodel.ChainCrl, 0)
			biggerFiles := make([]chainmodel.ChainCrl, 0)
			for i, sameAkiChainCrl := range sameAkiChainCrls {
				// using filename and crlnumber to found self ( may have same filename )
				if sameAkiChainCrl.FileName == chainCrl.FileName && sameAkiChainCrl.CrlNumber == chainCrl.CrlNumber {
					if i > 0 && i < len(sameAkiChainCrls) {
						smallerFiles = sameAkiChainCrls[:i]
					}
					if i+1 < len(sameAkiChainCrls) {
						biggerFiles = sameAkiChainCrls[i+1:]
					}

					belogs.Debug("validateCrl():same aki have crl files are smaller or bigger: self: i, aki, CrlNumber:",
						i, chainCrl.Aki, chainCrl.CrlNumber,
						",  sameAkiChainCrls are ", jsonutil.MarshalJson(sameAkiChainCrls),
						",  smallerFiles are ", jsonutil.MarshalJson(smallerFiles),
						",  biggerFiles files are ", jsonutil.MarshalJson(biggerFiles))

					if len(biggerFiles) == 0 {

						stateMsg := model.StateMsg{Stage: "chainvalidate",
							Fail: "There are multiple CRLs under a specific AKI, and this CRL has the largest CRL Number",
							Detail: "the smaller files are " + jsonutil.MarshalJson(smallerFiles) +
								", the bigge files are " + jsonutil.MarshalJson(biggerFiles)}
						//chainCrl.StateModel.AddWarning(&stateMsg)
						belogs.Debug("validateCrl():len(biggerFiles) == 0, all same aki crl files are smaller, so it is just warning, ",
							chainCrl.Aki, chainCrl.FileName, chainCrl.CrlNumber, "  sameAkiChainCrls: ", jsonutil.MarshalJson(sameAkiChainCrls), stateMsg)

					} else {

						stateMsg := model.StateMsg{Stage: "chainvalidate",
							Fail: "There are multiple CRLs under a specific AKI, and this CRL has not the largest CRL Number",
							Detail: "the smaller files are " + jsonutil.MarshalJson(smallerFiles) +
								", the bigge files are " + jsonutil.MarshalJson(biggerFiles)}
						chainCrl.StateModel.AddError(&stateMsg)
						belogs.Debug("validateCrl():len(biggerFiles) > 0, some same aki crl files are bigger, so it is error, ",
							chainCrl.Aki, chainCrl.FileName, chainCrl.CrlNumber, "  sameAkiChainCrls: ", jsonutil.MarshalJson(sameAkiChainCrls),
							"  bigger files:", jsonutil.MarshalJson(biggerFiles), stateMsg)

					}
					break
				}
			}
		}
	}
	chainCrl.StateModel.JudgeState()
	belogs.Debug("validateCrl(): stateModel:", chainCrl.StateModel)
	if chainCrl.StateModel.State != "valid" {
		belogs.Info("validateCrl(): stateModel have errors or warnings, crlId :", crlId, "  stateModel:", jsonutil.MarshalJson(chainCrl.StateModel))
	}
	chains.UpdateFileTypeIdToCrl(&chainCrl)
	belogs.Debug("validateCrl():end UpdateFileTypeIdToCrl crlId:", crlId, "  time(s):", time.Now().Sub(start).Seconds())

}
func getCrlParentChainCers(chains *chainmodel.Chains, crlId uint64) (chainCerAlones []chainmodel.ChainCerAlone, err error) {

	parentChainCerAlone, err := getCrlParentChainCer(chains, crlId)
	if err != nil {
		belogs.Error("getCrlParentChainCers(): getCrlParentChainCer, crlId:", crlId, err)
		return nil, err
	}
	belogs.Debug("getCrlParentChainCers(): crlId:", crlId, "  parentChainCerAlone.Id:", parentChainCerAlone.Id)

	if parentChainCerAlone.Id == 0 {
		belogs.Debug("getCrlParentChainCers(): parentChainCer is not found , crlId :", crlId)
		return chainCerAlones, nil
	}

	chainCerAlones = make([]chainmodel.ChainCerAlone, 0)
	chainCerAlones = append(chainCerAlones, parentChainCerAlone)
	chainCerAlonesTmp, err := GetCerParentChainCers(chains, parentChainCerAlone.Id)
	if err != nil {
		belogs.Error("getCrlParentChainCers(): GetCerParentChainCers, crlId:", crlId, "   parentChainCerAlone.Id:", parentChainCerAlone.Id, err)
		return nil, err
	}
	chainCerAlones = append(chainCerAlones, chainCerAlonesTmp...)
	belogs.Debug("getCrlParentChainCers():crlId, len(chainCerAlones):", crlId, len(chainCerAlones))
	return chainCerAlones, nil
}

func getCrlParentChainCer(chains *chainmodel.Chains, crlId uint64) (chainCerAlone chainmodel.ChainCerAlone, err error) {
	chainCrl, err := chains.GetCrlById(crlId)
	if err != nil {
		belogs.Error("getCrlParentChainCer(): GetCrl, crlId:", crlId, err)
		return chainCerAlone, err
	}
	belogs.Debug("getCrlParentChainCer(): crlId:", crlId, "  chainCrl.Id:", chainCrl.Id)

	//get mft's aki --> parent cer's ski
	if len(chainCrl.Aki) == 0 {
		belogs.Error("getCrlParentChainCer(): chainCrl.Aki is empty, fail:", crlId)
		return chainCerAlone, errors.New("crl's aki is empty")
	}
	aki := chainCrl.Aki
	parentCerSki := aki
	fileTypeId, ok := chains.SkiToFileTypeId[parentCerSki]
	belogs.Debug("getCrlParentChainCer():crlId, parentCerSki, fileTypeId, ok:", crlId, parentCerSki, fileTypeId, ok)
	if ok {
		parentChainCer, err := chains.GetCerByFileTypeId(fileTypeId)
		belogs.Debug("getCrlParentChainCer(): GetCerByFileTypeId, crlId, fileTypeId, parentChainCer.Id:", crlId, fileTypeId, parentChainCer.Id)
		if err != nil {
			belogs.Error("getCrlParentChainCer(): GetCerByFileTypeId, crlId,fileTypeId, fail:", crlId, fileTypeId, err)
			return chainCerAlone, err
		}
		return *chainmodel.NewChainCerAlone(&parentChainCer), nil

	}
	//  not found parent ,is not error
	belogs.Debug("getCrlParentChainCer(): not found crl's parent cer:", crlId)
	return chainCerAlone, nil
}
func getSameAkiChainCrls(chains *chainmodel.Chains, crlId uint64) (sameAkiChainCrls []chainmodel.ChainCrl, err error) {
	chainCrl, err := chains.GetCrlById(crlId)
	if err != nil {
		belogs.Error("getSameAkiChainCrls():GetCrlById, crlId:", crlId, err)
		return
	}

	sameAkiChainCrls = make([]chainmodel.ChainCrl, 0)
	//get crl's aki --> cer/roa/crl/
	aki := chainCrl.Aki
	fileTypeIds, ok := chains.AkiToFileTypeIds[aki]
	belogs.Debug("getSameAkiChainCrls():crlId,  fileTypeIds, ok:", crlId, fileTypeIds, ok)
	if ok {
		for _, fileTypeId := range fileTypeIds.FileTypeIds {
			belogs.Debug("getSameAkiChainCrls():crlId, fileTypeId:", crlId, fileTypeId)
			if ok {
				fileType := string(fileTypeId[:3])
				switch fileType {
				case "crl":
					chainCrl, err := chains.GetCrlByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiChainCrls(): GetCrlByFileTypeId, crlId,fileTypeId,err:", crlId, fileTypeId, err)
						return nil, err
					}
					sameAkiChainCrls = append(sameAkiChainCrls, chainCrl)
					belogs.Debug("getSameAkiChainCrls():crlId,  sameAkiChainCrls:", crlId, sameAkiChainCrls)
				}
			}
		}
	}
	return
}
