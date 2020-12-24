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
	hashutil "github.com/cpusoft/goutil/hashutil"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"

	"chainvalidate/db"
	chainmodel "chainvalidate/model"
	"model"
)

func GetChainMfts(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	belogs.Debug("GetChainMfts(): start:")

	chainMftSqls, err := db.GetChainMftSqls()
	if err != nil {
		belogs.Error("GetChainMfts(): db.GetChainMftSqls:", err)
		return
	}
	belogs.Debug("GetChainMfts(): GetChainMftSqls, len(chainMftSqls):", len(chainMftSqls))

	for i := range chainMftSqls {
		chainMft := chainMftSqls[i].ToChainMft()
		chainMft.ChainFileHashs, err = db.GetChainFileHashs(chainMft.Id)
		belogs.Debug("GetChainMfts():i, chainMft:", i, jsonutil.MarshalJson(chainMft))
		if err != nil {
			belogs.Error("GetChainCers(): db.GetChainFileHashs fail:", chainMft.Id, err)
			return
		}
		chains.MftIds = append(chains.MftIds, chainMftSqls[i].Id)
		chains.AddMft(&chainMft)
	}

	belogs.Debug("GetChainMfts(): end len(chainMftSqls):", len(chainMftSqls), ",   len(chains.MftIds):", len(chains.MftIds), "  time(s):", time.Now().Sub(start).Seconds())
	return
}

func ValidateMfts(chains *chainmodel.Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()

	mftIds := chains.MftIds
	belogs.Debug("ValidateMfts(): start: len(mftIds):", len(mftIds))

	var mftWg sync.WaitGroup
	chainMftCh := make(chan int, conf.Int("chain::chainConcurrentCount"))
	for _, mftId := range mftIds {
		mftWg.Add(1)
		chainMftCh <- 1
		go validateMft(chains, mftId, &mftWg, chainMftCh)
	}
	mftWg.Wait()
	close(chainMftCh)

	belogs.Info("ValidateMfts(): end, len(mftIds):", len(mftIds), "  time(s):", time.Now().Sub(start).Seconds())

}

func validateMft(chains *chainmodel.Chains, mftId uint64, wg *sync.WaitGroup, chainMftCh chan int) {
	defer func() {
		wg.Done()
		<-chainMftCh
	}()

	start := time.Now()
	chainMft, err := chains.GetMftById(mftId)
	if err != nil {
		belogs.Error("validateMft(): GetMftById fail:", mftId, err)
		return
	}
	// set parent cer
	chainMft.ParentChainCerAlones, err = getMftParentChainCers(chains, mftId)
	if err != nil {
		belogs.Error("validateMft(): getMftParentChainCers fail:", mftId, err)
		chainMft.StateModel.JudgeState()
		chains.UpdateFileTypeIdToMft(&chainMft)
		return
	}
	belogs.Debug("validateMft(): chainMft.ParentChainCer:  mftId,  len(chainMft.ParentChainCerAlones):", mftId, len(chainMft.ParentChainCerAlones))

	// exists parent cer
	if len(chainMft.ParentChainCerAlones) > 0 {
		// get one parent
		parentCer := osutil.JoinPathFile(chainMft.ParentChainCerAlones[0].FilePath, chainMft.ParentChainCerAlones[0].FileName)
		mft := osutil.JoinPathFile(chainMft.FilePath, chainMft.FileName)
		belogs.Debug("validateMft():parentCer:", parentCer, "    mft:", mft)

		// openssl verify mft
		result, err := certutil.VerifyEeCertByX509(parentCer, mft, chainMft.EeCertStart, chainMft.EeCertEnd)
		belogs.Debug("validateMft():VerifyEeCertByX509 result:", result, err)
		if result != "ok" {
			desc := ""
			if err != nil {
				desc = err.Error()
				belogs.Debug("validateMft():verify mft by parent cer fail, fail, mftId:", chainMft.Id, err)
			}
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to be verified by its issuing certificate",
				Detail: desc + ",  parent cer file is " + chainMft.ParentChainCerAlones[0].FileName + ",  mft file is " + chainMft.FileName}
			// if subject doesnot match ,will just set warning
			if strings.Contains(desc, "issuer name does not match subject from issuing certificate") {
				chainMft.StateModel.AddWarning(&stateMsg)
			} else {
				chainMft.StateModel.AddError(&stateMsg)
			}

		}
	} else {
		belogs.Debug("validateMft():mft file has not found parent cer, fail, chainMft.Id,mftId:", chainMft.Id, mftId)
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Its issuing certificate no longer exists",
			Detail: ""}
		chainMft.StateModel.AddError(&stateMsg)
	}

	// check files in filehash should exist
	noExistFiles := make([]string, 0)
	sha256ErrorFiles := make([]string, 0)
	for _, fh := range chainMft.ChainFileHashs {
		f := osutil.JoinPathFile(fh.Path, fh.File)
		exist, err := osutil.IsExists(f)
		belogs.Debug("validateMft():IsExists f:", f, exist, err)
		if !exist || err != nil {
			belogs.Debug("validateMft():IsExists f fail:", f, exist, err)
			noExistFiles = append(noExistFiles, fh.File)
			continue
		}

		sha256, err := hashutil.Sha256File(f)
		belogs.Debug("validateMft():Sha256File:", f, "  calc hash:"+sha256, " fh.Hash:"+fh.Hash)
		if sha256 != fh.Hash || err != nil {
			belogs.Debug("validateMft():Sha256File  fail,  mftfile is ", chainMft.FilePath+chainMft.FileName,
				" err fil is "+f,
				"  calc sha256:"+sha256, "  saved sha256:"+fh.Hash, err)
			sha256ErrorFiles = append(sha256ErrorFiles, f)
			continue
		}
	}
	if len(noExistFiles) > 0 {
		belogs.Debug("validateMft():verify mft file fail, mftId:", chainMft.Id,
			"   noExistFiles:", jsonutil.MarshalJson(noExistFiles))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail: "File on filelist no longer exists",
			Detail: "object(s) is(are) not in publication point but listed on mft, the(these) object(s) is(are) " +
				strings.Join(noExistFiles, ", ")}
		if conf.Bool("policy::allowInMftNoExist") {
			chainMft.StateModel.AddWarning(&stateMsg)
		} else {
			chainMft.StateModel.AddError(&stateMsg)
		}

	}
	if len(sha256ErrorFiles) > 0 {
		belogs.Debug("validateMft():verify mft file hash fail, mftId:", chainMft.Id,
			"   sha256ErrorFiles:", jsonutil.MarshalJson(sha256ErrorFiles))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail: "The sha256 value of the file is not equal to the value on the filelist",
			Detail: "object(s) in publication point and mft has(have) different hashvalues, the(these) object(s) is(are) " +
				strings.Join(sha256ErrorFiles, ", ")}
		if conf.Bool("policy::allowIncorrectMftHashValue") {
			chainMft.StateModel.AddWarning(&stateMsg)
		} else {
			chainMft.StateModel.AddError(&stateMsg)
		}
	}

	belogs.Debug("validateMft():after check ChainFileHashs, stateModel:", chainMft.Id, jsonutil.MarshalJson(chainMft.StateModel))

	noExistFiles = make([]string, 0)
	// check all the file(cer/crl/roa) which have same aki ,should all in filehash
	sameAkiCerRoaCrlFiles, sameAkiChainMfts, err := getSameAkiCerRoaCrlFilesChainMfts(chains, mftId)
	if err != nil {
		belogs.Debug("validateMft():GetSameAkiCerRoaCrlFiles fail, aki:", chainMft.Aki)
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Fail to get CER/ROA/CRL/MFT under specific AKI",
			Detail: err.Error()}
		chainMft.StateModel.AddError(&stateMsg)
	} else {

		if len(sameAkiCerRoaCrlFiles) == 0 {
			belogs.Debug("validateMft():GetSameAkiCerRoaCrlFiles len(akiFiles)==0, aki:", chainMft.Aki)
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to get CER/ROA/CRL/MFT under specific AKI",
				Detail: "the aki is " + chainMft.Aki}
			chainMft.StateModel.AddError(&stateMsg)
		}

		for _, sameAkiCerRoaCrlFile := range sameAkiCerRoaCrlFiles {
			found := false
			for _, fileHash := range chainMft.ChainFileHashs {
				if strings.ToLower(sameAkiCerRoaCrlFile) == strings.ToLower(fileHash.File) {
					found = true
					break
				}
			}
			if !found {
				belogs.Debug("validateMft():the same aki file ", sameAkiCerRoaCrlFile, " is not exist in filehashs of mft ")
				noExistFiles = append(noExistFiles, sameAkiCerRoaCrlFile)
			}
		}

		if len(noExistFiles) > 0 {
			belogs.Debug("validateMft():the same aki " + chainMft.Aki + " files " + jsonutil.MarshalJson(noExistFiles) + "  is not exists in filehashs of mft")
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail: "The CER, ROA and CRL of these same AKI are not on the filelist of MFT of same AKI",
				Detail: "object(s) is(are) in publication point but not listed on mft, the(these) object(s) is(are) " +
					jsonutil.MarshalJson(noExistFiles)}
			if conf.Bool("policy::allowCerRoaCrlNotInMft") {
				chainMft.StateModel.AddWarning(&stateMsg)
			} else {
				chainMft.StateModel.AddError(&stateMsg)
			}
		}
	}
	belogs.Debug("validateMft():after check akiFiles, stateModel:", chainMft.Id, jsonutil.MarshalJson(chainMft.StateModel))

	// check same aki mft files, compare mftnumber
	// mft files have only one
	belogs.Debug("validateMft():GetSameAkiMftFiles aki:", chainMft.Aki,
		" self is ", chainMft.FileName,
		" chainMfts:", jsonutil.MarshalJson(sameAkiChainMfts))
	if len(sameAkiChainMfts) == 1 {
		// filename shoud be equal
		if sameAkiChainMfts[0].FileName != chainMft.FileName {
			belogs.Debug("validateMft():same mft files is not self, aki:", sameAkiChainMfts[0].FileName, chainMft.FileName, chainMft.Aki)
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to get Manifest under specific AKI",
				Detail: "aki is" + chainMft.Aki + "  fileName is " + chainMft.FileName + "  same aki file is " + sameAkiChainMfts[0].FileName}
			chainMft.StateModel.AddError(&stateMsg)
		}
	} else if len(sameAkiChainMfts) == 0 {
		belogs.Debug("validateMft():same mft files is zero, aki:", chainMft.Aki)
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Fail to get Manifest under specific AKI",
			Detail: "aki is " + chainMft.Aki + ",  fileName should be " + chainMft.FileName}
		chainMft.StateModel.AddError(&stateMsg)
	} else {
		belogs.Debug("validateMft():more than one same aki mft files, ",
			chainMft.Aki, chainMft.FileName, chainMft.MftNumber, "  sameAkiChainMfts: ", jsonutil.MarshalJson(sameAkiChainMfts))
		// smaller/older are more ahead
		smallerFiles := make([]chainmodel.ChainMft, 0)
		biggerFiles := make([]chainmodel.ChainMft, 0)
		for i, sameAkiChainMft := range sameAkiChainMfts {
			// using filename and mftnumber to found self ( may have same filename )
			if sameAkiChainMft.FileName == chainMft.FileName && sameAkiChainMft.MftNumber == chainMft.MftNumber {
				if i > 0 && i < len(sameAkiChainMfts) {
					smallerFiles = sameAkiChainMfts[:i]
				}
				if i+1 < len(sameAkiChainMfts) {
					biggerFiles = sameAkiChainMfts[i+1:]
				}

				belogs.Debug("validateMft():same aki have mft files are smaller or bigger: self: i, aki, mftNumber:",
					i, chainMft.Aki, chainMft.MftNumber,
					",  mftFiles are ", jsonutil.MarshalJson(sameAkiChainMfts),
					",  smallerFiles are ", jsonutil.MarshalJson(smallerFiles),
					",  biggerFiles files are ", jsonutil.MarshalJson(biggerFiles))

				if len(biggerFiles) == 0 {
					stateMsg := model.StateMsg{Stage: "chainvalidate",
						Fail: "There are multiple CRLs under a specific AKI, and this CRL has the largest CRL Number",
						Detail: "the smaller files are " + jsonutil.MarshalJson(smallerFiles) +
							", the bigge files are " + jsonutil.MarshalJson(biggerFiles)}
					//chainMft.StateModel.AddWarning(&stateMsg)
					belogs.Debug("validateMft():len(biggerFiles) == 0, all same aki mft files are smaller, so it is just warning, ",
						chainMft.Aki, chainMft.FileName, chainMft.MftNumber, "  sameAkiChainMfts: ", jsonutil.MarshalJson(sameAkiChainMfts), stateMsg)

				} else {
					stateMsg := model.StateMsg{Stage: "chainvalidate",
						Fail: "There are multiple Manifests under a specific AKI, and this Manifest has not the largest Manifest Number",
						Detail: "the smaller files are " + jsonutil.MarshalJson(smallerFiles) +
							", the bigge files are " + jsonutil.MarshalJson(biggerFiles)}
					chainMft.StateModel.AddError(&stateMsg)
					belogs.Debug("validateMft():len(biggerFiles) > 0, some same aki mft files are bigger, so it is error, ",
						chainMft.Aki, chainMft.FileName, chainMft.MftNumber, "  sameAkiChainMfts: ", jsonutil.MarshalJson(sameAkiChainMfts),
						"  bigger files:", jsonutil.MarshalJson(biggerFiles), stateMsg)
				}
				break
			}
		}

	}

	if len(chainMft.ChainSnInCrlRevoked.CrlFileName) > 0 {
		belogs.Debug("validateMft(): mft ee file is founded in crl's revoked cer list:",
			chainMft.Id, jsonutil.MarshalJson(chainMft.ChainSnInCrlRevoked.CrlFileName))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail: "The EE of this Manifest is found on the revocation list of CRL",
			Detail: chainMft.FileName + " is in " + chainMft.ChainSnInCrlRevoked.CrlFileName + " revoked cer list, " +
				" and revoked time is " + convert.Time2StringZone(chainMft.ChainSnInCrlRevoked.RevocationTime)}
		chainMft.StateModel.AddError(&stateMsg)
	}

	chainMft.StateModel.JudgeState()
	belogs.Debug("validateMft(): stateModel:", chainMft.StateModel)
	if chainMft.StateModel.State != "valid" {
		belogs.Info("validateMft(): stateModel have errors or warnings, mftId :", mftId, "  stateModel:", jsonutil.MarshalJson(chainMft.StateModel))
	}
	chains.UpdateFileTypeIdToMft(&chainMft)
	belogs.Debug("validateMft():end UpdateFileTypeIdToMft  mftId:", mftId, "  time(s):", time.Now().Sub(start).Seconds())

}

func getMftParentChainCers(chains *chainmodel.Chains, mftId uint64) (chainCerAlones []chainmodel.ChainCerAlone, err error) {

	parentChainCerAlone, err := getMftParentChainCer(chains, mftId)
	if err != nil {
		belogs.Error("getMftParentChainCers(): getMftParentChainCer, mftId:", mftId, err)
		return nil, err
	}
	belogs.Debug("getMftParentChainCers(): mftId:", mftId, "  parentChainCerAlone.Id:", parentChainCerAlone.Id)

	if parentChainCerAlone.Id == 0 {
		belogs.Debug("getMftParentChainCers(): parentChainCer is not found , mftId :", mftId)
		return chainCerAlones, nil
	}

	chainCerAlones = make([]chainmodel.ChainCerAlone, 0)
	chainCerAlones = append(chainCerAlones, parentChainCerAlone)
	chainCerAlonesTmp, err := GetCerParentChainCers(chains, parentChainCerAlone.Id)
	if err != nil {
		belogs.Error("getMftParentChainCers(): GetCerParentChainCers, mftId:", mftId, "   parentChainCerAlone.Id:", parentChainCerAlone.Id, err)
		return nil, err
	}
	chainCerAlones = append(chainCerAlones, chainCerAlonesTmp...)
	belogs.Debug("getMftParentChainCers():mftId, len(chainCerAlones):", mftId, len(chainCerAlones))
	return chainCerAlones, nil
}

func getMftParentChainCer(chains *chainmodel.Chains, mftId uint64) (chainCerAlone chainmodel.ChainCerAlone, err error) {
	chainMft, err := chains.GetMftById(mftId)
	if err != nil {
		belogs.Error("getMftParentChainCer(): GetMft, mftId:", mftId, err)
		return chainCerAlone, err
	}
	belogs.Debug("getMftParentChainCer(): mftId:", mftId, "  chainMft:", chainMft)

	//get mft's aki --> parent cer's ski
	if len(chainMft.Aki) == 0 {
		belogs.Error("getMftParentChainCer(): chainMft.Aki is empty, fail:", mftId)
		return chainCerAlone, errors.New("mft's aki is empty")
	}

	aki := chainMft.Aki
	parentCerSki := aki
	fileTypeId, ok := chains.SkiToFileTypeId[parentCerSki]
	belogs.Debug("getMftParentChainCer(): mftId, parentCerSki,fileTypeId, ok:", mftId, parentCerSki, fileTypeId, ok)
	if ok {
		parentChainCer, err := chains.GetCerByFileTypeId(fileTypeId)
		belogs.Debug("getMftParentChainCer(): GetCerByFileTypeId, mftId, fileTypeId, parentChainCer.Id:", mftId, fileTypeId, parentChainCer.Id)
		if err != nil {
			belogs.Error("getMftParentChainCer(): GetCerByFileTypeId, mftId,fileTypeId, fail:", mftId, fileTypeId, err)
			return chainCerAlone, err
		}
		return *chainmodel.NewChainCerAlone(&parentChainCer), nil
	}
	//  not found parent ,is not error
	belogs.Debug("getMftParentChainCer(): not found mft's parent cer:", mftId)
	return chainCerAlone, nil
}

func getSameAkiCerRoaCrlFilesChainMfts(chains *chainmodel.Chains, mftId uint64) (sameAkiCerRoaCrlFiles []string,
	sameAkiChainMfts []chainmodel.ChainMft, err error) {
	chainMft, err := chains.GetMftById(mftId)
	if err != nil {
		belogs.Error("getSameAkiCerRoaCrlFilesChainMfts():GetMftById, mftId:", mftId, err)
		return
	}

	sameAkiCerRoaCrlFiles = make([]string, 0)
	sameAkiChainMfts = make([]chainmodel.ChainMft, 0)
	//get mft's aki --> cer/roa/crl/
	aki := chainMft.Aki
	fileTypeIds, ok := chains.AkiToFileTypeIds[aki]
	belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId, fileTypeIds, ok:", mftId, fileTypeIds, ok)
	if ok {
		for _, fileTypeId := range fileTypeIds.FileTypeIds {
			belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId, fileTypeId:", mftId, fileTypeId)
			if ok {
				fileType := string(fileTypeId[:3])
				switch fileType {
				case "cer":
					chainCer, err := chains.GetCerByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetCerByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
						return nil, nil, err
					}
					sameAkiCerRoaCrlFiles = append(sameAkiCerRoaCrlFiles, chainCer.FileName)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId, chainCer.FileName, ok:", mftId, chainCer.FileName, ok)
				case "crl":
					chainCrl, err := chains.GetCrlByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetCrlByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
						return nil, nil, err
					}
					sameAkiCerRoaCrlFiles = append(sameAkiCerRoaCrlFiles, chainCrl.FileName)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId, chainCrl.FileName, ok:", mftId, chainCrl.FileName, ok)
				case "roa":
					chainRoa, err := chains.GetRoaByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetRoaByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
						return nil, nil, err
					}
					sameAkiCerRoaCrlFiles = append(sameAkiCerRoaCrlFiles, chainRoa.FileName)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId,  chainRoa.FileName, ok:", mftId, chainRoa.FileName, ok)
				case "mft":
					chainMft, err := chains.GetMftByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetMftByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
						return nil, nil, err
					}
					sameAkiChainMfts = append(sameAkiChainMfts, chainMft)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId, chainMft.Id, ok:", mftId, chainMft.Id, ok)
				}
			}
		}
	}
	return
}
