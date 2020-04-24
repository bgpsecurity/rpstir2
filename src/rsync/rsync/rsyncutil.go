package rsync

import (
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	randutil "github.com/cpusoft/goutil/randutil"
	rsyncutil "github.com/cpusoft/goutil/rsyncutil"

	"model"
)

func RsyncResult2LabRpkiSyncLogFile(rsyncResult *rsyncutil.RsyncResult, labRpkiSyncLogId uint64, body string) model.LabRpkiSyncLogFile {
	belogs.Debug("RsyncResult2LabRpkiSyncLogFile():rsyncResult:", jsonutil.MarshalJson(rsyncResult), "    labRpkiSyncLogId:", labRpkiSyncLogId)
	labRpkiSyncLogFile := model.LabRpkiSyncLogFile{}
	labRpkiSyncLogFile.FileName = rsyncResult.FileName
	labRpkiSyncLogFile.FilePath = rsyncResult.FilePath
	labRpkiSyncLogFile.FileType = rsyncResult.FileType
	labRpkiSyncLogFile.SyncLogId = labRpkiSyncLogId
	labRpkiSyncLogFile.ParseValidateResultJson = string(body)
	labRpkiSyncLogFile.SyncTime = rsyncResult.SyncTime
	labRpkiSyncLogFile.SyncType = rsyncResult.RsyncType

	// only roa need set to rtr. this is initial set. more detail set will after parse to set
	rtr := "notNeed"
	if rsyncResult.FileType == "roa" {
		rtr = "notYet"
	}

	state := model.LabRpkiSyncLogFileState{
		Sync:            "finished",
		UpdateCertTable: "notYet",
		Rtr:             rtr,
	}
	labRpkiSyncLogFile.State = jsonutil.MarshalJson(state)
	belogs.Debug("RsyncResult2LabRpkiSyncLogFile():convert rsyncResult to labRpkiSyncLogFile:", jsonutil.MarshalJson(labRpkiSyncLogFile))
	return labRpkiSyncLogFile
}

func AddSubCaRepositoryUrlsToRpQueue(subCaRepositoryUrls []string) {

	rsyncConcurrentCount := conf.Int("rsync::rsyncConcurrentCount")
	belogs.Debug("AddSubCaRepositoryUrlsToRpQueue(): len(rpQueue.RsyncModelChan)+len(subCaRepositoryUrls):", len(rpQueue.RsyncModelChan),
		" + ", len(subCaRepositoryUrls), " compare rsync::rsyncConcurrentCount ", rsyncConcurrentCount)
	for i, subCaRepositoryUrl := range subCaRepositoryUrls {
		belogs.Debug("AddSubCaRepositoryUrlsToRpQueue():waitForRsyncUrl, rpQueue.RsyncingParsingCount: ",
			atomic.LoadInt64(&rpQueue.RsyncingParsingCount),
			"   subCaRepositoryUrl:", subCaRepositoryUrl)
		if !rpQueue.PreCheckRsyncUrl(subCaRepositoryUrl) {
			belogs.Debug("AddSubCaRepositoryUrlsToRpQueue():PreCheckRsyncUrl before RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
			atomic.AddInt64(&rpQueue.RsyncingParsingCount, -1)
			belogs.Debug("AddSubCaRepositoryUrlsToRpQueue():PreCheckRsyncUrl after RsyncingParsingCount-1:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
			continue
		}

		curRsyncingCount := int(atomic.LoadInt64(&rpQueue.CurRsyncingCount))
		if curRsyncingCount <= 2 {
			// when less 2, not need to wait
		} else if curRsyncingCount > 2 && curRsyncingCount <= rsyncConcurrentCount {
			waitForRsyncUrl(1, subCaRepositoryUrl)
		} else {
			belogs.Debug("AddSubCaRepositoryUrlsToRpQueue():waitForRsyncUrl,i + rpQueue.curRsyncingCount: ", curRsyncingCount)
			waitForRsyncUrl(i+curRsyncingCount, subCaRepositoryUrl)
		}
		go rpQueue.AddRsyncUrl(subCaRepositoryUrl, conf.VariableString("rsync::destpath")+"/")
	}
}

// rsync concurrent is more than rsync::rsyncConcurrentCount, should wait;;;
// willAddRsyncCount is len(subCaRepositoryUrls, will add len(rpQueue.RsyncModelChan), than compare to  conf.Int("rsync::rsyncConcurrentCount")
func WaitForRsyncParseEnd() bool {
	waitCount := int(rpQueue.RsyncMisc.OkRsyncUrlLen) + len(rpQueue.RsyncMisc.FailRsyncUrls)
	belogs.Debug("WaitForRsyncParseEnd():waitCount: ",
		rpQueue.RsyncMisc.OkRsyncUrlLen, len(rpQueue.RsyncMisc.FailRsyncUrls))
	for i := 0; i < waitCount*2; i++ {
		belogs.Debug("WaitForRsyncParseEnd():waitCount rpQueue.RsyncingParsingCount,i: ",
			atomic.LoadInt64(&rpQueue.RsyncingParsingCount), i)
		if atomic.LoadInt64(&rpQueue.RsyncingParsingCount) != 0 {
			belogs.Debug("WaitForRsyncParseEnd(): return false:   rpQueue.RsyncingParsingCount:",
				atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
			return false
		} else {
			//will check again
			time.Sleep(time.Duration(10) * time.Millisecond)
		}
	}
	belogs.Debug("WaitForRsyncParseEnd(): return true, will end ")
	return true
}

// will try fail urls  to rsync again
func TryAgainFailRsyncUrls() bool {
	// try again
	belogs.Debug("TryAgainFailRsyncUrls():try fail urls again: len(rpQueue.RsyncMisc.FailRsyncUrls):", len(rpQueue.RsyncMisc.FailRsyncUrls),
		"      rpQueue.RsyncMisc.FailRsyncUrlsTryCount:", rpQueue.RsyncMisc.FailRsyncUrlsTryCount)
	if len(rpQueue.RsyncMisc.FailRsyncUrls) > 0 &&
		rpQueue.RsyncMisc.FailRsyncUrlsTryCount <= uint64(conf.Int("rsync::failRsyncUrlsTryCount")) {
		failRsyncUrls := make([]string, 0, len(rpQueue.RsyncMisc.FailRsyncUrls))
		for failRsyncUrl := range rpQueue.RsyncMisc.FailRsyncUrls {
			failRsyncUrls = append(failRsyncUrls, failRsyncUrl)
			// delete saved url ,so can try again
			rpQueue.DelRsyncAddedUrl(failRsyncUrl)
		}
		// clear fail rsync urls
		rpQueue.RsyncMisc.FailRsyncUrls = make(map[string]string, 200)

		belogs.Debug("TryAgainFailRsyncUrls(): failRysncUrl:", len(failRsyncUrls), failRsyncUrls,
			"   rpQueue.RsyncMisc.FailRsyncUrlsTryCount: ", rpQueue.RsyncMisc.FailRsyncUrlsTryCount)
		atomic.AddUint64(&rpQueue.RsyncMisc.FailRsyncUrlsTryCount, 1)
		belogs.Debug("TryAgainFailRsyncUrls():after  rpQueue.RsyncMisc.FailRsyncUrlsTryCount: ", rpQueue.RsyncMisc.FailRsyncUrlsTryCount)

		// check rsync concurrent count, wait some time,
		rsyncConcurrentCount := conf.Int("rsync::rsyncConcurrentCount")
		atomic.AddInt64(&rpQueue.RsyncingParsingCount, int64(len(failRsyncUrls)))
		belogs.Debug("TryAgainFailRsyncUrls(): len(failRsyncUrls):", len(failRsyncUrls),
			"  failRsyncUrls:", failRsyncUrls)

		for i, failRsyncUrl := range failRsyncUrls {
			curRsyncingCount := int(atomic.LoadInt64(&rpQueue.CurRsyncingCount))
			if curRsyncingCount <= 2 {

			} else if curRsyncingCount > 2 && curRsyncingCount <= rsyncConcurrentCount {
				belogs.Debug("TryAgainFailRsyncUrls():waitForRsyncUrl, i is smaller, i: ", i,
					" , will wait  1:", 1)
				waitForRsyncUrl(1, failRsyncUrl)
			} else {
				belogs.Debug("TryAgainFailRsyncUrls():waitForRsyncUrl, i is bigger, i: ", i,
					" , will wait  curRsyncingCount+1:", curRsyncingCount+1)
				waitForRsyncUrl(1+curRsyncingCount/2, failRsyncUrl)
			}
			go rpQueue.AddRsyncUrl(failRsyncUrl, conf.VariableString("rsync::destpath")+"/")
		}
		return true
	}
	return false
}

//rsync  should wait for some url, because some nic limited access frequency
func waitForRsyncUrl(curRsyncCount int, url string) {

	if curRsyncCount == 0 {
		return
	}
	belogs.Debug("waitForRsyncUrl(): curRsyncCount : ", curRsyncCount, "  will add:", conf.Int("rsync::rsyncDefaultWaitMs"), " 2* runtime.NumCPU():", 2*runtime.NumCPU())
	curRsyncCount = curRsyncCount + conf.Int("rsync::rsyncDefaultWaitMs") + 2*runtime.NumCPU()

	// apnic and afrinic should not visit too often
	if strings.Contains(url, "rpki.apnic.net") {
		curRsyncCount = curRsyncCount * 2
	} else if strings.Contains(url, "rpki.afrinic.net") {
		curRsyncCount = curRsyncCount * 10
	}
	min := uint(conf.Int("rsync::rsyncPerDelayMs") * curRsyncCount)
	randR := uint(conf.Int("rsync::rsyncDelayRandMs"))
	rand := randutil.IntRange(min, randR)
	belogs.Debug("waitForRsyncUrl():after rand, url is :", url, ",  curRsyncCount is:", curRsyncCount, ", will sleep rand ms:", rand)
	time.Sleep(time.Duration(rand) * time.Millisecond)
}
