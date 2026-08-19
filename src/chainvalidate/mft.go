package chainvalidate

import (
	"errors"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/certutil"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

func getChainMfts(chains *Chains, chainWg *sync.WaitGroup, syncLogId uint64) {

	defer chainWg.Done()
	start := time.Now()

	var mftWg sync.WaitGroup
	chainMftDatasCh := make(chan []*ChainCertData, conf.Int("chain::getDbConcurrentCount"))
	go callAddMftToChain(chains, syncLogId, chainMftDatasCh, &mftWg)
	err := GetChainMftData(chainMftDatasCh, &mftWg)
	if err != nil {
		belogs.Error("getChainMfts(): GetChainMftData fail: ", err)
		close(chainMftDatasCh)
		return
	}
	belogs.Debug("getChainMfts(): after GetChainMftData before chainWg.Wait(), time(s):", time.Since(start))
	mftWg.Wait()
	close(chainMftDatasCh)
	belogs.Info("getChainMfts(): end, syncLogId:", syncLogId, " time(s):", time.Since(start))
}

func callAddMftToChain(chains *Chains, syncLogId uint64,
	chainMftDatasCh chan []*ChainCertData, mftWg *sync.WaitGroup) {
	start := time.Now()
	var index uint64
	for {
		select {
		case chainMftDatas, ok := <-chainMftDatasCh:
			belogs.Debug("callAddMftToChain(): get from chainMftDatasCh, len(chainMftDatas):", len(chainMftDatas),
				"  index:", index, "  ok:", ok)
			if ok {
				go addMftToChain(chains, chainMftDatas, syncLogId, mftWg)
				index++
				belogs.Debug("callAddMftToChain(): addMftToChain, syncLogId:", syncLogId, " index:", index)
			} else {
				belogs.Info("callAddMftToChain(): close chainMftSqlsCh, index:", index, "   time(s):", time.Since(start))
				return
			}
		}
	}
}

func addMftToChain(chains *Chains, chainMftDatas []*ChainCertData,
	syncLogId uint64, mftWg *sync.WaitGroup) {
	defer func() {
		mftWg.Done()
		belogs.Debug("addMftToChain(): mftWg.Done(), len(chainMftDatas):", len(chainMftDatas))
	}()

	start := time.Now()
	for i := range chainMftDatas {
		startOne := time.Now()
		belogs.Debug("addMftToChain(): chainMftDatas:", jsonutil.MarshalJson(chainMftDatas[i]), "  syncLogId:", syncLogId)

		var err error
		chainMft, err := chainMftDatas[i].ToChainMft()
		if err != nil {
			belogs.Error("addMftToChain(): ToChainMft fail, chainMftSql.Id:", chainMftDatas[i].Id, err)
			return
		}

		belogs.Debug("addMftToChain(): will getChainFileHashsDb, id:", chainMft.Id)

		chainMft.ChainFileHashs, err = GetChainFileHashs(chainMft)
		if err != nil {
			belogs.Error("addMftToChain(): getChainFileHashsDb fail, id:", chainMft.Id, err, "  time(s):", time.Since(start))
			return
		}
		belogs.Debug("addMftToChain(): len(chainMft.ChainFileHashs):", len(chainMft.ChainFileHashs), " id:", chainMft.Id, "  time(s):", time.Since(start))

		// todo 需要获取
		chainMft.PreviousMft, err = GetPreviousMft(chainMft)
		belogs.Debug("addMftToChain(): previousMft:", chainMft.PreviousMft)
		if err != nil {
			belogs.Error("addMftToChain(): getPreviousMftDb fail, id:", chainMft.Id, err, "  time(s):", time.Since(start))
			return
		}
		belogs.Debug("addMftToChain(): chainMft.PreviousMft:", jsonutil.MarshalJson(chainMft.PreviousMft), " id:", chainMft.Id, "  time(s):", time.Since(start)) //shaodebug

		chains.AddMftId(chainMftDatas[i].Id)
		chains.AddMft(&chainMft)
		belogs.Debug("addMftToChain(): added, chainMft:", jsonutil.MarshalJson(chainMft),
			"  syncLogId:", syncLogId, "    chainMftDatas.SyncLogId:", chainMftDatas[i].SyncLogId,
			"  time(s):", time.Since(startOne))
	}
	belogs.Info("addMftToChain(): all added, len(chainMftDatas):", len(chainMftDatas), "  time(s):", time.Since(start))
	return
}

func validateMfts(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()

	mftIds := chains.MftIds
	belogs.Debug("validateMfts(): start: len(mftIds):", len(mftIds))

	var mftWg sync.WaitGroup
	chainMftCh := make(chan int, conf.Int("chain::validateConcurrentCount"))
	for _, mftId := range mftIds {
		mftWg.Add(1)
		chainMftCh <- 1
		go validateMft(chains, mftId, &mftWg, chainMftCh)
	}
	mftWg.Wait()
	close(chainMftCh)

	belogs.Info("validateMfts(): validate end, len(mftIds):", len(mftIds), "  time(s):", time.Since(start))

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
	belogs.Debug("validateMft(): GetMftById, mftId:", mftId, "  file:", chainMft.FilePath, chainMft.FileName, "  time(s):", time.Since(start))

	// set parent cer
	chainMft.ParentChainCerAlones, err = getMftParentChainCers(chains, mftId)
	if err != nil {
		belogs.Error("validateMft(): getMftParentChainCers fail:", mftId, err)
		chainMft.StateModel.JudgeState()
		chains.UpdateFileTypeIdToMft(&chainMft)
		return
	}
	belogs.Debug("validateMft(): getMftParentChainCers, mftId:", mftId, "  file:", chainMft.FilePath, chainMft.FileName,
		"  len(chainMft.ParentChainCerAlones):", len(chainMft.ParentChainCerAlones), "  time(s):", time.Since(start))

	// exists parent cer
	if len(chainMft.ParentChainCerAlones) > 0 {
		// get one parent
		parentCer := osutil.JoinPathFile(chainMft.ParentChainCerAlones[0].FilePath, chainMft.ParentChainCerAlones[0].FileName)
		mft := osutil.JoinPathFile(chainMft.FilePath, chainMft.FileName)
		belogs.Debug("validateMft():parentCer:", parentCer, "    mft:", mft)

		// openssl verify mft
		result, err := certutil.VerifyEeCertByX509(parentCer, mft, chainMft.EeCertStart, chainMft.EeCertEnd)
		belogs.Debug("validateMft():VerifyEeCertByX509 result:", result, err, "  time(s):", time.Since(start))
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
		belogs.Debug("validateMft():IsExists mftId:", mftId, "  mftfile:", chainMft.FilePath+chainMft.FileName,
			"  file:", f, "  exist:", exist, err)
		if err != nil || !exist {
			belogs.Error("validateMft():IsExists file fail, mftId:", mftId, "  mftfile:", chainMft.FilePath+chainMft.FileName,
				"  file:", f, "  exist:", exist, err)
			noExistFiles = append(noExistFiles, fh.File)
			continue
		}

		sha256, err := hashutil.Sha256File(f)
		belogs.Debug("validateMft():Sha256File mftId:", mftId, "  mftfile:", chainMft.FilePath+chainMft.FileName,
			"  file:", f, "  calc hash:"+sha256, " fh.Hash:"+fh.Hash, err)
		if err != nil || sha256 != fh.Hash {
			belogs.Error("validateMft():Sha256File file, fail,  mftId:", mftId, "  mftfile:", chainMft.FilePath+chainMft.FileName,
				"  err file is "+f,
				"  calc sha256:"+sha256, "  saved sha256:"+fh.Hash, err)
			sha256ErrorFiles = append(sha256ErrorFiles, f)
			continue
		}
	}
	belogs.Debug("validateMft(): check ChainFileHashs, mftId:", mftId, "  mftfile:", chainMft.FilePath, chainMft.FileName,
		"  time(s):", time.Since(start))
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
	belogs.Debug("validateMft():after check ChainFileHashs,  mftId:", mftId, "  file:", chainMft.FilePath, chainMft.FileName,
		"  stateModel:", jsonutil.MarshalJson(chainMft.StateModel), "  time(s):", time.Since(start))

	noExistFiles = make([]string, 0)
	// check all the file(cer/crl/roa) which have same aki ,should all in filehash
	sameAkiCerRoaAsaCrlFiles, sameAkiCrls, sameAkiChainMfts, err := getSameAkiCerRoaCrlFilesChainMfts(chains, mftId)
	belogs.Debug("validateMft():getSameAkiCerRoaCrlFilesChainMfts,mftId:", mftId, "  file:", chainMft.FilePath, chainMft.FileName,
		"   sameAkiCerRoaAsaCrlFiles:", sameAkiCerRoaAsaCrlFiles,
		"   sameAkiCrls:", sameAkiCrls, "   sameAkiChainMfts:", sameAkiChainMfts, err, "  time(s):", time.Since(start))
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
					Fail: "The ThisUpdate of CRL is different from the ThisUpdate of MFT which has the same AKI",
					Detail: "The ThisUpdate of CRL is " + convert.ToString(sameAkiCrls[i].ThisUpdate) +
						", and the ThisUpdate of MFT is " + convert.ToString(chainMft.ThisUpdate) +
						", and the CLR file is " + sameAkiCrls[i].FilePath + " " + sameAkiCrls[i].FileName}
				chainMft.StateModel.AddWarning(&stateMsg)
			}
			if !chainMft.NextUpdate.Equal(sameAkiCrls[i].NextUpdate) {
				stateMsg := model.StateMsg{Stage: "chainvalidate",
					Fail: "The NextUpdate of CRL is different from the NextUpdate of MFT which has same AKI",
					Detail: "The NextUpdate of CRL is " + convert.ToString(sameAkiCrls[i].NextUpdate) +
						", and the NextUpdate of MFT is " + convert.ToString(chainMft.ThisUpdate) +
						", and the CLR file is " + sameAkiCrls[i].FilePath + " " + sameAkiCrls[i].FileName}
				chainMft.StateModel.AddWarning(&stateMsg)
			}
		}

	}
	belogs.Debug("validateMft():after check akiFiles, mftId:", mftId, "  file:", chainMft.FilePath, chainMft.FileName,
		" stateModel:", chainMft.Id, jsonutil.MarshalJson(chainMft.StateModel), "  time(s):", time.Since(start))

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
						Fail: "There are multiple Manifests under a specific AKI, and this Manifest has the largest Manifest Number",
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
	belogs.Debug("validateMft():after check sameAkiChainMfts, mftId:", mftId, "  file:", chainMft.FilePath, chainMft.FileName,
		" stateModel:", jsonutil.MarshalJson(chainMft.StateModel), "  time(s):", time.Since(start))

	if len(chainMft.ChainSnInCrlRevoked.CrlFileName) > 0 {
		belogs.Debug("validateMft(): mft ee file is founded in crl's revoked cer list:",
			chainMft.Id, jsonutil.MarshalJson(chainMft.ChainSnInCrlRevoked.CrlFileName))
		stateMsg := model.StateMsg{Stage: "chainvalidate",
			Fail: "The EE of this Manifest is found on the revocation list of CRL",
			Detail: chainMft.FileName + " is in " + chainMft.ChainSnInCrlRevoked.CrlFileName + " revoked cer list, " +
				" and revoked time is " + convert.Time2StringZone(chainMft.ChainSnInCrlRevoked.RevocationTime)}
		chainMft.StateModel.AddError(&stateMsg)
	}
	belogs.Debug("validateMft(): after check ChainSnInCrlRevoked,  mftId:", mftId, "  file:", chainMft.FilePath, chainMft.FileName,
		"   chainMft.PreviousMft:", chainMft.PreviousMft, "  time(s):", time.Since(start))

	if chainMft.PreviousMft.Found {
		// compare prev Number and cur NUmber
		prevMftNumber, okPrev := new(big.Int).SetString(chainMft.PreviousMft.MftNumber, 16)
		curMftNumber, ok := new(big.Int).SetString(chainMft.MftNumber, 16)
		// shaodebug
		belogs.Debug("validateMft(): found previous mft,  mftId:", chainMft.Id,
			"   prevMftNumber:", prevMftNumber, "   okPrev:", okPrev, "   curMftNumber:", curMftNumber, "  ok:", ok)
		// should be hex
		if !ok || !okPrev {
			belogs.Info("validateMft(): !ok || !okPrev   mftId:", chainMft.Id)
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
		belogs.Debug("validateMft(): prevMftNumber and curMftNumber,  mftId:", mftId, "  file:", chainMft.FilePath, chainMft.FileName,
			"  chainMft.StateModel:", jsonutil.MarshalJson(chainMft.StateModel), "  time(s):", time.Since(start))

		// compare prev thisUpdate/nextUpdate and cur thisUpdate/nextUpdate
		if !chainMft.ThisUpdate.After(chainMft.PreviousMft.ThisUpdate) {
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "The ThisUpdate of this Manifest is not later than the previous ThisUpdate",
				Detail: "The ThisUpdate of this Manifest is " + chainMft.ThisUpdate.String() + ", and the previouse ThisUpdate is " + chainMft.PreviousMft.ThisUpdate.String()}
			chainMft.StateModel.AddError(&stateMsg)
		}
		if !chainMft.NextUpdate.After(chainMft.PreviousMft.NextUpdate) {
			stateMsg := model.StateMsg{Stage: "chainvalidate",
				Fail:   "The NextUpdate of this Manifest is not later than the previous NextUpdate",
				Detail: "The NextUpdate of this Manifest is " + chainMft.NextUpdate.String() + ", and the previouse NextUpdate is " + chainMft.PreviousMft.NextUpdate.String()}
			chainMft.StateModel.AddError(&stateMsg)
		}
		belogs.Debug("validateMft(): ThisUpdate and NextUpdate,   mftId:", chainMft.Id, "  chainMft.StateModel:", jsonutil.MarshalJson(chainMft.StateModel)) //shaodebug
	}
	belogs.Debug("validateMft(): after check PreviousMft,  mftId:", mftId, "  file:", chainMft.FilePath, chainMft.FileName,
		"   chainMft.PreviousMft:", chainMft.PreviousMft, "  time(s):", time.Since(start))

	chainMft.StateModel.JudgeState()
	belogs.Debug("validateMft(): stateModel:", chainMft.StateModel)
	if chainMft.StateModel.State != "valid" {
		belogs.Debug("validateMft(): stateModel have errors or warnings, mftId :", mftId, "  stateModel:", jsonutil.MarshalJson(chainMft.StateModel))
	}
	chains.UpdateFileTypeIdToMft(&chainMft)
	belogs.Info("validateMft(): validate one end, UpdateFileTypeIdToMft, mftId:", mftId,
		"  mft file:", chainMft.FilePath, chainMft.FileName, "  time(s):", time.Since(start))

}

func getMftParentChainCers(chains *Chains, mftId uint64) (chainCerAlones []ChainCerAlone, err error) {

	start := time.Now()
	parentChainCerAlone, err := getMftParentChainCer(chains, mftId)
	if err != nil {
		belogs.Error("getMftParentChainCers(): getMftParentChainCer, mftId:", mftId, err)
		return nil, err
	}
	belogs.Debug("getMftParentChainCers(): mftId:", mftId, "  parentChainCerAlone.Id:", parentChainCerAlone.Id, "  time(s):", time.Since(start))

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
	belogs.Debug("getMftParentChainCers():mftId:", mftId, " len(chainCerAlones):", len(chainCerAlones), "  time(s):", time.Since(start))
	return chainCerAlones, nil
}

func getMftParentChainCer(chains *Chains, mftId uint64) (chainCerAlone ChainCerAlone, err error) {
	start := time.Now()
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
	belogs.Debug("getMftParentChainCer(): SkiToFileTypeId, mftId:", mftId, "  parentCerSki:", parentCerSki, "  fileTypeId:", fileTypeId, "  ok:", ok)
	if ok {
		parentChainCer, err := chains.GetCerByFileTypeId(fileTypeId)
		belogs.Debug("getMftParentChainCer(): GetCerByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId,
			"  parentChainCer.Id:", parentChainCer.Id, "   time(s):", time.Since(start))
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
	start := time.Now()
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
	belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): AkiToFileTypeIds, mftId:", mftId, "  fileTypeIds:", fileTypeIds, "  ok:", ok)
	if ok {
		for _, fileTypeId := range fileTypeIds.FileTypeIds {
			belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): range FileTypeIds, mftId:", mftId, "  fileTypeId:", fileTypeId)
			if ok {
				fileType := string(fileTypeId[:3])
				switch fileType {
				case "cer":
					chainCer, err := chains.GetCerByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetCerByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId, err)
						return nil, nil, nil, err
					}
					sameAkiCerRoaAsaCrlFiles = append(sameAkiCerRoaAsaCrlFiles, chainCer.FileName)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): GetCerByFileTypeId mftId:", mftId, " chainCer.FileName:", chainCer.FileName)
				case "crl":
					chainCrl, err := chains.GetCrlByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetCrlByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId, err)
						return nil, nil, nil, err
					}
					sameAkiCerRoaAsaCrlFiles = append(sameAkiCerRoaAsaCrlFiles, chainCrl.FileName)
					sameAkiCrl := SameAkiCrl{Found: true,
						FilePath:   chainCrl.FilePath,
						FileName:   chainCrl.FileName,
						ThisUpdate: chainCrl.ThisUpdate,
						NextUpdate: chainCrl.NextUpdate}
					sameAkiCrls = append(sameAkiCrls, sameAkiCrl)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): GetCrlByFileTypeId mftId:", mftId, " chainCrl.FileName:", chainCrl.FileName, "  sameAkiCrl:", sameAkiCrl)
				case "roa":
					chainRoa, err := chains.GetRoaByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetRoaByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId, err)
						return nil, nil, nil, err
					}
					sameAkiCerRoaAsaCrlFiles = append(sameAkiCerRoaAsaCrlFiles, chainRoa.FileName)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): mftId:", mftId, "  chainRoa.FileName:", chainRoa.FileName)
				case "asa":
					chainAsa, err := chains.GetAsaByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetAsaByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId, err)
						return nil, nil, nil, err
					}
					sameAkiCerRoaAsaCrlFiles = append(sameAkiCerRoaAsaCrlFiles, chainAsa.FileName)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): GetAsaByFileTypeId mftId:", mftId, "  chainAsa.FileName:", chainAsa.FileName)
				case "mft":
					chainMft, err := chains.GetMftByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetMftByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId, err)
						return nil, nil, nil, err
					}
					sameAkiChainMfts = append(sameAkiChainMfts, chainMft)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): GetMftByFileTypeId mftId:", mftId, "  chainMft.FileName:", chainMft.FileName)
				case "moa":
					chainMoa, err := chains.GetMoaByFileTypeId(fileTypeId)
					if err != nil {
						belogs.Error("getSameAkiCerRoaCrlFilesChainMfts(): GetMoaByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId, err)
						return nil, nil, nil, err
					}
					sameAkiCerRoaAsaCrlFiles = append(sameAkiCerRoaAsaCrlFiles, chainMoa.FileName)
					belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): GetMoaByFileTypeId mftId:", mftId, "  chainMoa.FileName:", chainMoa.FileName)
				}
			}
		}
	}
	belogs.Debug("getSameAkiCerRoaCrlFilesChainMfts(): get all, mftId:", mftId, "  time(s):", time.Since(start))
	return
}

// invalidMftEffect:warning/invalid
func updateChainByMft(chains *Chains, invalidMftEffect string) (err error) {
	start := time.Now()
	mftIds := chains.MftIds
	belogs.Info("updateChainByMft(): start: len(mftIds):", len(mftIds))
	rsyncDestPath := conf.String("rsync::destPath") + "/"
	rrdpDestPath := conf.String("rrdp::destPath") + "/"
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
					belogs.Error("updateChainByMft(): GetCerByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId, err)
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
					belogs.Error("updateChainByMft(): GetCrlByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId, err)
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
					belogs.Error("updateChainByMft(): GetRoaByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId, err)
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
					belogs.Error("updateChainByMft(): GetAsaByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId, err)
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
			case "moa":
				chainMoa, err := chains.GetMoaByFileTypeId(fileTypeId)
				if err != nil {
					belogs.Error("updateChainByMft(): GetMoaByFileTypeId, mftId:", mftId, "  fileTypeId:", fileTypeId, err)
					return err
				}
				if invalidMftEffect == "warning" {
					chainMoa.StateModel.AddWarning(&stateMsg)
				} else if invalidMftEffect == "invalid" {
					chainMoa.StateModel.AddError(&stateMsg)
				}
				chains.UpdateFileTypeIdToMoa(&chainMoa)
				belogs.Debug("updateChainByMft(): mftId:", mftId, "   chainMft:", chainMft.FilePath, chainMft.FileName,
					" chainMoa:", chainMoa.FilePath, chainMoa.FileName, jsonutil.MarshalJson(chainMoa.StateModel))
			default:
				// do nothing
			}

		}

	}
	belogs.Info("updateChainByMft(): end: len(mftIds):", len(mftIds), "  time(s):", time.Since(start))
	return nil
}

func updateMfts(chains *Chains, wg *sync.WaitGroup) {
	defer wg.Done()

	start := time.Now()
	err := UpdateMfts(chains)
	if err != nil {
		belogs.Error("updateMfts(): UpdateMfts fail:", err)
		return
	}
	belogs.Info("updateMfts(): ok, time(s):", time.Since(start))
}
