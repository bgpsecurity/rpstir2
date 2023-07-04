package chainvalidate

import (
	"errors"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/certutil"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	model "rpstir2-model"
)

func getChainMfts(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	belogs.Debug("getChainMfts(): start:")

	chainMftSqls, err := getChainMftSqlsDb()
	if err != nil {
		belogs.Error("getChainMfts(): getChainMftSqlsDb:", err)
		return
	}
	belogs.Debug("getChainMfts(): getChainMftSqlsDb, len(chainMftSqls):", len(chainMftSqls))

	for i := range chainMftSqls {
		chainMft := chainMftSqls[i].ToChainMft()
		chainMft.ChainFileHashs, err = GetChainFileHashsDb(chainMft.Id)
		if err != nil {
			belogs.Error("getChainMfts(): GetChainFileHashsDb fail:", chainMft.Id, err)
			return
		}
		belogs.Debug("getChainMfts():i:", i, " chainMft.ChainFileHashs:", chainMft.ChainFileHashs)

		chainMft.PreviousMft, err = GetPreviousMftDb(chainMft.Id)
		belogs.Debug("getChainMfts():i:", i, " previousMft:", i, chainMft.PreviousMft)
		if err != nil {
			belogs.Error("getChainMfts(): GetPreviousMftDb fail:", chainMft.Id, err)
			return
		}
		belogs.Debug("getChainMfts():i:", i, " chainMft.PreviousMft:", chainMft.PreviousMft) //shaodebug

		chains.MftIds = append(chains.MftIds, chainMftSqls[i].Id)
		chains.AddMft(&chainMft)
	}

	belogs.Debug("getChainMfts(): end len(chainMftSqls):", len(chainMftSqls), ",   len(chains.MftIds):", len(chains.MftIds), "  time(s):", time.Since(start))
	return
}

func validateMfts(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()

	mftIds := chains.MftIds
	belogs.Debug("validateMfts(): start: len(mftIds):", len(mftIds))

	var mftWg sync.WaitGroup
	chainMftCh := make(chan int, conf.Int("chain::chainConcurrentCount"))
	for _, mftId := range mftIds {
		mftWg.Add(1)
		chainMftCh <- 1
		go validateMft(chains, mftId, &mftWg, chainMftCh)
	}
	mftWg.Wait()
	close(chainMftCh)

	belogs.Info("validateMfts(): end, len(mftIds):", len(mftIds), "  time(s):", time.Since(start))

}

func validateMft(chains *Chains, mftId uint64, wg *sync.WaitGroup, chainMftCh chan int) {
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
		chainMft.StateModel.AddError(&stateMsg)
	}
	if len(sha256ErrorFiles) > 0 {
		belogs.Debug("validateMft():verify mft file hash fail, mftId:", chainMft.Id,
			"   sha256ErrorFiles:", jsonutil.MarshalJson(sha256ErrorFiles))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail: "The sha256 value of the file is not equal to the value on the filelist",
			Detail: "object(s) in publication point and mft has(have) different hashvalues, the(these) object(s) is(are) " +
				strings.Join(sha256ErrorFiles, ", ")}
		chainMft.StateModel.AddError(&stateMsg)

	}
	belogs.Debug("validateMft():after check ChainFileHashs, stateModel:", chainMft.Id, jsonutil.MarshalJson(chainMft.StateModel))

	noExistFiles = make([]string, 0)
	// check all the file(cer/crl/roa) which have same aki ,should all in filehash
	sameAkiCerRoaAsaCrlFiles, sameAkiCrls, sameAkiChainMfts, err := getSameAkiCerRoaCrlFilesChainMfts(chains, mftId)
	belogs.Debug("validateMft():getSameAkiCerRoaCrlFilesChainMfts, mftId:", chainMft.Id, "   sameAkiCerRoaAsaCrlFiles:", sameAkiCerRoaAsaCrlFiles,
		"   sameAkiCrls:", sameAkiCrls, "   sameAkiChainMfts:", sameAkiChainMfts, err)
	if err != nil {
		belogs.Debug("validateMft():getSameAkiCerRoaCrlFilesChainMfts fail, aki:", chainMft.Aki)
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Fail to get CER/ROA/CRL/MFT under specific AKI",
			Detail: err.Error()}
		chainMft.StateModel.AddError(&stateMsg)
	} else {

		if len(sameAkiCerRoaAsaCrlFiles) == 0 {
			belogs.Debug("validateMft():getSameAkiCerRoaCrlFilesChainMfts len(akiFiles)==0, aki:", chainMft.Aki)
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to get CER/ROA/CRL/MFT under specific AKI",
				Detail: "the aki is " + chainMft.Aki}
			chainMft.StateModel.AddError(&stateMsg)
		}

		for _, sameAkiCerRoaCrlFile := range sameAkiCerRoaAsaCrlFiles {
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
			chainMft.StateModel.AddError(&stateMsg)
		}

		// mft's thisUpdate/nextUpdate are equal to clr's thisUpdate/nextUpdate
		if len(sameAkiCrls) == 0 {
			belogs.Debug("validateMft():getSameAkiCerRoaCrlFilesChainMfts len(sameAkiCrls)==0, aki:", chainMft.Aki)
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "Fail to get CRL under specific AKI",
				Detail: "The aki of MFT is " + chainMft.Aki}
			chainMft.StateModel.AddError(&stateMsg)
		}
		for i := range sameAkiCrls {
			if !chainMft.ThisUpdate.Equal(sameAkiCrls[i].ThisUpdate) {
				stateMsg := model.StateMsg{Stage: "chainvalidate",
					Fail: "The thisUpdate of CRL is different from thisUpdate of MFT which has the same AKI",
					Detail: "The thisUpdate of CRL is " + convert.ToString(sameAkiCrls[i].ThisUpdate) +
						", and the thisUpdate of MFT is " + convert.ToString(chainMft.ThisUpdate) +
						", and the CLR file is " + sameAkiCrls[i].FilePath + " " + sameAkiCrls[i].FileName}
				chainMft.StateModel.AddWarning(&stateMsg)
			}
			if !chainMft.NextUpdate.Equal(sameAkiCrls[i].NextUpdate) {
				stateMsg := model.StateMsg{Stage: "chainvalidate",
					Fail: "The nextUpdate of CRL is different from nextUpdate of MFT which has same AKI",
					Detail: "The NextUpdate of CRL is " + convert.ToString(sameAkiCrls[i].NextUpdate) +
						", and the NextUpdate of MFT is " + convert.ToString(chainMft.ThisUpdate) +
						", and the CLR file is " + sameAkiCrls[i].FilePath + " " + sameAkiCrls[i].FileName}
				chainMft.StateModel.AddWarning(&stateMsg)
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
				Detail: "aki is " + chainMft.Aki + "  fileName is " + chainMft.FileName + "  same aki file is " + sameAkiChainMfts[0].FileName}
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
		smallerFiles := make([]ChainMft, 0)
		biggerFiles := make([]ChainMft, 0)
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

	belogs.Debug("validateMft(): check previous mft,  mftId:", chainMft.Id, "   chainMft.PreviousMft:", chainMft.PreviousMft) //shaodebug
	if chainMft.PreviousMft.Found {
		// compare prev Number and cur NUmber
		prevMftNumber, okPrev := new(big.Int).SetString(chainMft.PreviousMft.MftNumber, 16)
		curMftNumber, ok := new(big.Int).SetString(chainMft.MftNumber, 16)
		// shaodebug
		belogs.Debug("validateMft(): found previous mft,  mftId:", chainMft.Id,
			"   prevMftNumber:", prevMftNumber, "   okPrev:", okPrev, "   curMftNumber:", curMftNumber, "  ok:", ok)
		// should be hex
		if !ok || !okPrev {
			belogs.Info("validateMft(): !ok || !okPrev   mftId:", chainMft.Id) //shaodebug
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "The Number of this Manifest or the previous Number is not a Hexadecimal number",
				Detail: "The Number of this Manifest is " + chainMft.MftNumber + ", and the previouse Number is " + chainMft.PreviousMft.MftNumber}
			chainMft.StateModel.AddError(&stateMsg)
		} else {

			comp := curMftNumber.Cmp(prevMftNumber)
			belogs.Debug("validateMft(): comp, prevMftNumber:", prevMftNumber, "   curMftNumber:", curMftNumber, "  comp:", comp) //shaodebug
			if comp < 0 {
				// if cur < prev, then error
				stateMsg := model.StateMsg{Stage: "chainvalidate",
					Fail:   "The Number of this Manifest is less than the previous Number",
					Detail: "The Number of this Manifest is " + curMftNumber.String() + ", and the previouse Number is " + prevMftNumber.String()}
				chainMft.StateModel.AddError(&stateMsg)
			} else if comp == 0 {
				// if cur == prev, then warning
				stateMsg := model.StateMsg{Stage: "chainvalidate",
					Fail:   "The Number of this Manifest is equal to the previous Number",
					Detail: "The Number of this Manifest is " + curMftNumber.String() + ", and the previouse Number is " + prevMftNumber.String()}
				chainMft.StateModel.AddWarning(&stateMsg)
			} else {
				// cur > prev
				// if cur - prev == 1 ,then ok, else warning
				one := big.NewInt(1)
				sub := big.NewInt(0).Sub(curMftNumber, prevMftNumber)
				belogs.Debug("validateMft(): comp, one:", one, "   sub:", sub) //shaodebug
				// just bigger 1, ok
				if sub.Cmp(one) != 0 {
					stateMsg := model.StateMsg{Stage: "chainvalidate",
						Fail:   "The Number of this Manifest is not exactly 1 larger than the previous Number",
						Detail: "The Number of this Manifest is " + curMftNumber.String() + ", and the previouse Number is " + prevMftNumber.String()}
					chainMft.StateModel.AddWarning(&stateMsg)
				}
			}
		}
		belogs.Debug("validateMft(): prevMftNumber and curMftNumber,   mftId:", chainMft.Id, "  chainMft.StateModel:", jsonutil.MarshalJson(chainMft.StateModel)) //shaodebug

		// compare prev thisUpdate/nextUpdate and cur thisUpdate/nextUpdate
		if !chainMft.ThisUpdate.After(chainMft.PreviousMft.ThisUpdate) {
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "The ThisUpdate of this Manifest is is later than the previous ThisUpdate",
				Detail: "The ThisUpdate of this Manifest is " + chainMft.ThisUpdate.String() + ", and the previouse ThisUpdate is " + chainMft.PreviousMft.ThisUpdate.String()}
			chainMft.StateModel.AddError(&stateMsg)
		}
		if !chainMft.NextUpdate.After(chainMft.PreviousMft.NextUpdate) {
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "The NextUpdate of this Manifest is is later than the previous NextUpdate",
				Detail: "The NextUpdate of this Manifest is " + chainMft.NextUpdate.String() + ", and the previouse NextUpdate is " + chainMft.PreviousMft.NextUpdate.String()}
			chainMft.StateModel.AddError(&stateMsg)
		}
		belogs.Debug("validateMft(): ThisUpdate and NextUpdate,   mftId:", chainMft.Id, "  chainMft.StateModel:", jsonutil.MarshalJson(chainMft.StateModel)) //shaodebug

	}

	chainMft.StateModel.JudgeState()
	belogs.Debug("validateMft(): stateModel:", chainMft.StateModel)
	if chainMft.StateModel.State != "valid" {
		belogs.Debug("validateMft(): stateModel have errors or warnings, mftId :", mftId, "  stateModel:", jsonutil.MarshalJson(chainMft.StateModel))
	}
	chains.UpdateFileTypeIdToMft(&chainMft)
	belogs.Debug("validateMft():end UpdateFileTypeIdToMft  mftId:", mftId, "  time(s):", time.Since(start))

}

func getMftParentChainCers(chains *Chains, mftId uint64) (chainCerAlones []ChainCerAlone, err error) {

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

	chainCerAlones = make([]ChainCerAlone, 0)
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

func getMftParentChainCer(chains *Chains, mftId uint64) (chainCerAlone ChainCerAlone, err error) {
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
		return *NewChainCerAlone(&parentChainCer), nil
	}
	//  not found parent ,is not error
	belogs.Debug("getMftParentChainCer(): not found mft's parent cer:", mftId)
	return chainCerAlone, nil
}

func getSameAkiCerRoaCrlFilesChainMfts(chains *Chains, mftId uint64) (sameAkiCerRoaAsaCrlFiles []string, sameAkiCrls []SameAkiCrl,
	sameAkiChainMfts []ChainMft, err error) {
	chainMft, err := chains.GetMftById(mftId)
	if err != nil {
		belogs.Error("getSameAkiCerRoaCrlFilesChainMfts():GetMftById, mftId:", mftId, err)
		return nil, nil, nil, err
	}

	sameAkiCerRoaAsaCrlFiles = make([]string, 0)
	sameAkiCrls = make([]SameAkiCrl, 0)
	sameAkiChainMfts = make([]ChainMft, 0)
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
						return nil, nil, nil, err
					}
					sameAkiCerRoaAsaCrlFiles = append(sameAkiCerRoaAsaCrlFiles, chainCer.FileName)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId, chainCer.FileName, ok:", mftId, chainCer.FileName, ok)
				case "crl":
					chainCrl, err := chains.GetCrlByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetCrlByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
						return nil, nil, nil, err
					}
					sameAkiCerRoaAsaCrlFiles = append(sameAkiCerRoaAsaCrlFiles, chainCrl.FileName)
					sameAkiCrl := SameAkiCrl{Found: true,
						FilePath:   chainCrl.FilePath,
						FileName:   chainCrl.FileName,
						ThisUpdate: chainCrl.ThisUpdate,
						NextUpdate: chainCrl.NextUpdate}
					sameAkiCrls = append(sameAkiCrls, sameAkiCrl)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId, chainCrl.FileName, ok:", mftId, chainCrl.FileName, ok, "  sameAkiCrl:", sameAkiCrl)
				case "roa":
					chainRoa, err := chains.GetRoaByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetRoaByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
						return nil, nil, nil, err
					}
					sameAkiCerRoaAsaCrlFiles = append(sameAkiCerRoaAsaCrlFiles, chainRoa.FileName)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId,  chainRoa.FileName, ok:", mftId, chainRoa.FileName, ok)
				case "asa":
					chainAsa, err := chains.GetAsaByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetAsaByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
						return nil, nil, nil, err
					}
					sameAkiCerRoaAsaCrlFiles = append(sameAkiCerRoaAsaCrlFiles, chainAsa.FileName)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId,  chainRoa.FileName, ok:", mftId, chainAsa.FileName, ok)
				case "mft":
					chainMft, err := chains.GetMftByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetMftByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
						return nil, nil, nil, err
					}
					sameAkiChainMfts = append(sameAkiChainMfts, chainMft)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId, chainMft.Id, ok:", mftId, chainMft.Id, ok)
				}
			}
		}
	}
	return
}

// invalidMftEffect:warning/invalid
func updateChainByMft(chains *Chains, invalidMftEffect string) (err error) {
	start := time.Now()
	mftIds := chains.MftIds
	belogs.Info("updateChainByMft(): start: len(mftIds):", len(mftIds))
	rsyncDestPath := conf.VariableString("rsync::destPath") + "/"
	rrdpDestPath := conf.VariableString("rrdp::destPath") + "/"
	// found invalid mft
	for _, mftId := range mftIds {
		chainMft, err := chains.GetMftById(mftId)
		if err != nil {
			belogs.Error("validateMft(): GetMftById fail:", mftId, err)
			return err
		}
		if chainMft.StateModel.State != "invalid" {
			continue
		}
		belogs.Debug("updateChainByMft(): found invalid mft, mftId:", mftId,
			chainMft.FilePath, chainMft.FileName, jsonutil.MarshalJson(chainMft.StateModel))
		fileTypeIds, ok := chains.AkiToFileTypeIds[chainMft.Aki]
		belogs.Debug("updateChainByMft(): mftId, fileTypeIds, ok:", mftId, fileTypeIds, ok)
		if !ok {
			continue
		}
		publicPointName := chainMft.FilePath
		publicPointName = strings.Replace(publicPointName, rsyncDestPath, "", -1)
		publicPointName = strings.Replace(publicPointName, rrdpDestPath, "", -1)

		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail:   "Manifest which has same AKI of this file is invalid or missing",
			Detail: `No manifest(invalid or missing) is available for ` + publicPointName + ` , and AKI is (` + chainMft.Aki + `), thus there may have been undetected deletions or replay substitutions from the publication point`}
		belogs.Debug("updateChainByMft(): mftId, publicPointName, stateMsg:", mftId, publicPointName,
			jsonutil.MarshalJson(stateMsg))

		for _, fileTypeId := range fileTypeIds.FileTypeIds {
			belogs.Debug("updateChainByMft(): mftId, fileTypeId:", mftId, fileTypeId)

			fileType := string(fileTypeId[:3])
			switch fileType {
			case "cer":
				chainCer, err := chains.GetCerByFileTypeId(fileTypeId)
				if err != nil {
					belogs.Error("updateChainByMft(): GetCerByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
					return err
				}
				if invalidMftEffect == "warning" {
					chainCer.StateModel.AddWarning(&stateMsg)
				} else if invalidMftEffect == "invalid" {
					chainCer.StateModel.AddError(&stateMsg)
				}
				chains.UpdateFileTypeIdToCer(&chainCer)
				belogs.Debug("updateChainByMft(): mftId:", mftId, "   chainMft:", chainMft.FilePath, chainMft.FileName,
					" chainCer:", chainCer.FilePath, chainCer.FileName, jsonutil.MarshalJson(chainCer.StateModel))
			case "crl":
				chainCrl, err := chains.GetCrlByFileTypeId(fileTypeId)
				if err != nil {
					belogs.Error("updateChainByMft(): GetCrlByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
					return err
				}
				if invalidMftEffect == "warning" {
					chainCrl.StateModel.AddWarning(&stateMsg)
				} else if invalidMftEffect == "invalid" {
					chainCrl.StateModel.AddError(&stateMsg)
				}
				chains.UpdateFileTypeIdToCrl(&chainCrl)
				belogs.Debug("updateChainByMft(): mftId:", mftId, "   chainMft:", chainMft.FilePath, chainMft.FileName,
					" chainCrl:", chainCrl.FilePath, chainCrl.FileName, jsonutil.MarshalJson(chainCrl.StateModel))
			case "roa":
				chainRoa, err := chains.GetRoaByFileTypeId(fileTypeId)
				if err != nil {
					belogs.Error("updateChainByMft(): GetRoaByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
					return err
				}
				if invalidMftEffect == "warning" {
					chainRoa.StateModel.AddWarning(&stateMsg)
				} else if invalidMftEffect == "invalid" {
					chainRoa.StateModel.AddError(&stateMsg)
				}
				chains.UpdateFileTypeIdToRoa(&chainRoa)
				belogs.Debug("updateChainByMft(): mftId:", mftId, "   chainMft:", chainMft.FilePath, chainMft.FileName,
					" chainRoa:", chainRoa.FilePath, chainRoa.FileName, jsonutil.MarshalJson(chainRoa.StateModel))
			case "asa":
				chainAsa, err := chains.GetAsaByFileTypeId(fileTypeId)
				if err != nil {
					belogs.Error("updateChainByMft(): GetAsaByFileTypeId, mftId,fileTypeId,err:", mftId, fileTypeId, err)
					return err
				}
				if invalidMftEffect == "warning" {
					chainAsa.StateModel.AddWarning(&stateMsg)
				} else if invalidMftEffect == "invalid" {
					chainAsa.StateModel.AddError(&stateMsg)
				}
				chains.UpdateFileTypeIdToAsa(&chainAsa)
				belogs.Debug("updateChainByMft(): mftId:", mftId, "   chainMft:", chainMft.FilePath, chainMft.FileName,
					" chainAsa:", chainAsa.FilePath, chainAsa.FileName, jsonutil.MarshalJson(chainAsa.StateModel))
			default:
				// do nothing
			}

		}

	}
	belogs.Info("updateChainByMft(): end: len(mftIds):", len(mftIds), "  time(s):", time.Since(start))
	return nil
}
