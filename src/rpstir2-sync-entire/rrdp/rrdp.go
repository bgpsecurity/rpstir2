package rrdp

import (
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	model "rpstir2-model"
	"rpstir2-sync-core/rrdp"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/randutil"
)

var rrQueue *RrdpParseQueue

// start to rrdp
func rrdpStart(syncUrls *model.SyncUrls) {
	belogs.Info("rrdpStart(): rrdp: syncUrls:", jsonutil.MarshalJson(syncUrls))

	syncRrdpLogs, err := rrdp.GetLastSyncRrdpLogsDb()
	if err != nil {
		belogs.Error("rrdpStart(): rrdp: GetLastSyncRrdpLogsDb fail:", err)
		return
	}

	//start rrQueue and rrdpForSelect
	rrQueue = NewQueue()
	rrQueue.LastSyncRrdpLogs = syncRrdpLogs
	rrQueue.LabRpkiSyncLogId = syncUrls.SyncLogId
	belogs.Debug("rrdpStart(): before startRrdpServer rrQueue:", jsonutil.MarshalJson(rrQueue))

	go startRrdpServer()
	belogs.Debug("rrdpStart(): after startRrdpServer rrQueue:", jsonutil.MarshalJson(rrQueue))

	// start to rrdp by sync url in tal, to get root cer
	// first: remove all root cer, so can will rrdp download and will trigger parse all cer files.
	// otherwise, will have to load all root file manually
	os.RemoveAll(conf.VariableString("rrdp::destPath") + "/root/")
	os.MkdirAll(conf.VariableString("rrdp::destPath")+"/root/", os.ModePerm)
	atomic.AddInt64(&rrQueue.RrdpingParsingCount, int64(len(syncUrls.RrdpUrls)))
	belogs.Debug("rrdpStart():after RrdpingParsingCount:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
	for _, url := range syncUrls.RrdpUrls {
		go rrQueue.AddRrdpUrl(url, conf.VariableString("rrdp::destPath")+"/")
	}
}

// start server ,wait input channel
func startRrdpServer() {
	start := time.Now()
	belogs.Info("startRrdpServer():start")

	for {
		select {
		case rrdpModelChan := <-rrQueue.RrdpModelChan:
			belogs.Debug("startRrdpServer(): rrdpModelChan:", rrdpModelChan,
				"  len(rrdprpQueue.RrdpModelChan):", len(rrQueue.RrdpModelChan),
				"  receive rrdpModelChan rrQueue.RrdpingParsingCount:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
			go rrdpByUrl(rrdpModelChan)
		case parseModelChan := <-rrQueue.ParseModelChan:
			belogs.Debug("startRrdpServer(): parseModelChan:", parseModelChan,
				"  receive parseModelChan rrQueue.RrdpingParsingCount:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
			go parseCerFiles(parseModelChan)
		case rrdpParseEndChan := <-rrQueue.RrdpParseEndChan:
			belogs.Debug("startRrdpServer():rrdpParseEndChan:", rrdpParseEndChan, "  rrQueue.RrdpingParsingCount:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))

			// try again the fail urls
			//belogs.Debug("startRrdpServer():try fail urls again: rrQueue.RrdpResult.FailRrdpUrls:", jsonutil.MarshalJson(rrQueue.RrdpResult.FailUrls))
			//if tryAgainFailRrdpUrls() {
			//		belogs.Debug("startRrdpServer(): tryAgainFailRrdpUrls continue")
			//		continue
			//	}
			rrQueue.RrdpResult.EndTime = time.Now()
			rrQueue.RrdpResult.OkUrls = rrQueue.GetRrdpUrls()
			rrQueue.RrdpResult.OkUrlsLen = uint64(len(rrQueue.RrdpResult.OkUrls))
			rrdpResultJson := jsonutil.MarshalJson(rrQueue.RrdpResult)
			belogs.Debug("startRrdpServer():end this rrdp success: rrdpResultJson:", rrdpResultJson)
			// will call sync to return result
			go func(rrdpResultJson string) {
				belogs.Debug("startRrdpServer():call /entiresync/rrdpresult: rrdpResultJson:", rrdpResultJson)
				httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
					"/entiresync/rrdpresult", rrdpResultJson, false)
			}(rrdpResultJson)

			// close rrQueue
			if !rrQueue.IsClose() {
				rrQueue.Close()
			}

			// return out of the for
			belogs.Info("startRrdpServer():end this rrdp success: rrdpResultJson:", rrdpResultJson, "  time(s):", time.Since(start))
			return
		}
	}
}

func rrdpByUrl(rrdpModelChan RrdpModelChan) {
	defer func() {
		belogs.Debug("RrdpByUrl():defer rrQueue.RrdpingParsingCount:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount),
			"  rrdpurls:", rrQueue.GetRrdpUrls())
		if atomic.LoadInt64(&rrQueue.RrdpingParsingCount) == 0 {
			belogs.Info("RrdpByUrl(): call RrdpParseEndChan{}, RrdpingParsingCount is 0")
			rrQueue.RrdpParseEndChan <- RrdpParseEndChan{}
		}
	}()

	// start rrdp and check err
	// if have error, should set RrdpingParsingCount -1
	start := time.Now()

	belogs.Debug("RrdpByUrl(): before rrdp, rrdpModelChan:", rrdpModelChan,
		"    CurRrdpingCount:", atomic.LoadInt64(&rrQueue.CurRrdpingCount), "   startTime:", start)
	lastSyncRrdpLog, hasLast := rrQueue.LastSyncRrdpLogs[rrdpModelChan.Url]
	rrdpByUrlModel := rrdp.RrdpByUrlModel{
		NotifyUrl:     rrdpModelChan.Url,
		DestPath:      rrdpModelChan.Dest,
		HasPath:       true,
		HasLast:       hasLast,
		LastSessionId: lastSyncRrdpLog.SessionId,
		LastCurSerial: lastSyncRrdpLog.CurSerial,
		SyncLogId:     rrQueue.LabRpkiSyncLogId,
	}
	belogs.Debug("RrdpByUrl():rrdpByUrlModel:", jsonutil.MarshalJson(rrdpByUrlModel))
	// will ignore connectRrdpUrlCh
	connectRrdpUrlCh := make(chan bool, 1)

	// CurRrdpingCount should +1 and then -1
	atomic.AddInt64(&rrQueue.CurRrdpingCount, 1)
	rrdpFiles, err := rrdp.RrdpByUrlImpl(rrdpByUrlModel, connectRrdpUrlCh, nil)
	atomic.AddInt64(&rrQueue.CurRrdpingCount, -1)
	belogs.Debug("RrdpByUrl(): RrdpByUrlImpl, len(rrdpFiles), err:", len(rrdpFiles), err)

	if err != nil {
		rrQueue.RrdpResult.FailUrls.Store(rrdpModelChan.Url, err.Error())
		belogs.Error("RrdpByUrl():RrdpByUrlImpl fail, rrdpModelChan.Url:", rrdpModelChan.Url, "   err:", err, "  time(s):", time.Since(start))
		belogs.Debug("RrdpByUrl():RrdpByUrlImpl fail, before RrdpingParsingCount-1:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
		atomic.AddInt64(&rrQueue.RrdpingParsingCount, -1)
		belogs.Debug("RrdpByUrl():RrdpByUrlImpl fail, after RrdpingParsingCount-1:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
		return
	}
	if len(rrdpFiles) == 0 {
		belogs.Debug("RrdpByUrl():len(rrdpFiles) == 0,no need rrdp,before:", rrdpModelChan.Url, "   , before RrdpingParsingCount-1:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
		atomic.AddInt64(&rrQueue.RrdpingParsingCount, -1)
		belogs.Debug("RrdpByUrl():len(rrdpFiles) == 0, no need rrdp,after:", rrdpModelChan.Url, "   , after RrdpingParsingCount-1:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
		return
	}
	belogs.Info("RrdpByUrl(): rrdp.Rrdp url:", rrdpModelChan.Url, "   len(rrdpFiles):", len(rrdpFiles))

	filePathNames := make([]string, 0)
	for i := range rrdpFiles {
		if osutil.ExtNoDot(rrdpFiles[i].FileName) == "cer" &&
			(rrdpFiles[i].SyncType == "add" || rrdpFiles[i].SyncType == "update") {
			filePathName := osutil.JoinPathFile(rrdpFiles[i].FilePath, rrdpFiles[i].FileName)
			filePathNames = append(filePathNames, filePathName)
		}
		if rrdpFiles[i].SyncType == "add" {
			atomic.AddUint64(&rrQueue.RrdpResult.AddFilesLen, 1)
		} else if rrdpFiles[i].SyncType == "del" {
			atomic.AddUint64(&rrQueue.RrdpResult.DelFilesLen, 1)
		} else if rrdpFiles[i].SyncType == "update" {
			atomic.AddUint64(&rrQueue.RrdpResult.UpdateFilesLen, 1)
		}

		belogs.Debug("RrdpByUrl(): rrdpFiles[i]:", jsonutil.MarshalJson(rrdpFiles[i]))
	}
	parseModelChan := ParseModelChan{FilePathNames: filePathNames}
	belogs.Debug("RrdpByUrl():parseModelChan:", parseModelChan, "   len(rrQueue.ParseModelChan):", len(rrQueue.ParseModelChan))
	belogs.Info("RrdpByUrl():  CurRrdpingCount:", atomic.LoadInt64(&rrQueue.CurRrdpingCount),
		"    len(filePathNames):", len(filePathNames),
		"    len(rrdpFiles):", len(rrdpFiles),
		//	"    rrQueue.RrdpResult:", jsonutil.MarshalJson(rrQueue.RrdpResult),
		"    time(s):", time.Since(start))

	rrQueue.ParseModelChan <- parseModelChan

}

func parseCerFiles(parseModelChan ParseModelChan) {
	defer func() {
		belogs.Debug("parseCerFiles():defer rrQueue.RrdpingParsingCount:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount),
			"  rrdpurls:", rrQueue.GetRrdpUrls())
		if atomic.LoadInt64(&rrQueue.RrdpingParsingCount) == 0 {
			belogs.Info("parseCerFiles(): call RyncParseEndChan{}, RrdpingParsingCount is 0")
			rrQueue.RrdpParseEndChan <- RrdpParseEndChan{}
		}
	}()
	belogs.Info("parseCerFiles(): parseModelChan:", parseModelChan.FilePathNames)

	// if have erorr, should set RrdpingParsingCount -1
	// get all cer files, include subcer

	// foreach every cerfiles to parseCerAndGetSubCaRepositoryUrl
	rpkiNotifies := make([]string, 0)
	for i := range parseModelChan.FilePathNames {
		// just trigger sync ,no need save to db
		rpkiNotify := parseCerAndGetRpkiNotify(parseModelChan.FilePathNames[i])
		if len(rpkiNotify) > 0 {
			rpkiNotifies = append(rpkiNotifies, rpkiNotify)
		} else {
			belogs.Info("parseCerFiles(): this file has no rpkiNotify:", parseModelChan.FilePathNames[i])
		}
	}

	// need sub this cer file, so -1
	var rrdpingParsingCountSub int64
	if len(rpkiNotifies) == 0 {
		rrdpingParsingCountSub = -1
	} else {
		rrdpingParsingCountSub = int64(len(rpkiNotifies)) - 1
	}
	belogs.Debug("parseCerFiles(): FilePathNames:", parseModelChan.FilePathNames,
		"   rpkiNotifies:", rpkiNotifies,
		"   rrdpingParsingCountSub:", rrdpingParsingCountSub)
	belogs.Info("parseCerFiles(): len(FilePathNames):", len(parseModelChan.FilePathNames),
		"   len(rpkiNotifies):", len(rpkiNotifies),
		"   rrdpingParsingCountSub:", rrdpingParsingCountSub)

	// the father rrdpingparsingcount -1 ,and the children rrdpingparsingcount + len()
	belogs.Debug("parseCerFiles():will add rpkiNotifies, before RrdpingParsingCount:",
		atomic.LoadInt64(&rrQueue.RrdpingParsingCount), "  rrdpingParsingCountSub:", rrdpingParsingCountSub)
	atomic.AddInt64(&rrQueue.RrdpingParsingCount, rrdpingParsingCountSub)
	belogs.Debug("parseCerFiles():will add rpkiNotifies, after RrdpingParsingCount:",
		atomic.LoadInt64(&rrQueue.RrdpingParsingCount), "  rrdpingParsingCountSub:", rrdpingParsingCountSub)

	// call add notifies to rrpqueue
	if len(rpkiNotifies) > 0 {
		addRpkiNotifiesToRpQueue(rpkiNotifies)
	}
}

// call /parsevalidate/parse to parse cert, and save result
func parseCerAndGetRpkiNotify(cerFile string) (rpkiNotify string) {

	// call parse, not need to save body to db
	start := time.Now()
	parseCerSimple := model.ParseCerSimple{}
	belogs.Debug("parseCerAndGetRpkiNotify():/parsevalidate/parsefilesimple cerFile:", cerFile)
	// post file, still use http
	err := httpclient.PostFileAndUnmarshalResponseModel("http://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpPort")+
		"/parsevalidate/parsefilesimple", cerFile, "file", false, &parseCerSimple)
	if err != nil {
		rrQueue.RrdpResult.FailParseValidateCerts.Store(cerFile, err.Error())
		belogs.Error("parseCerAndGetRpkiNotify(): PostFileAndUnmarshalResponseModel failed:", cerFile, "   err:", err)
		return ""
	}

	// get the sub repo url in cer, and send it to rpqueue
	belogs.Info("parseCerAndGetRpkiNotify(): cerFile:", cerFile, "    parseCerSimple:", jsonutil.MarshalJson(parseCerSimple),
		"  time(s):", time.Since(start))
	return parseCerSimple.RpkiNotify

}

func addRpkiNotifiesToRpQueue(rpkiNotifies []string) {

	belogs.Debug("addRpkiNotifiesToRpQueue(): len(rrQueue.RrdpModelChan)+len(subCaRepositoryUrls):", len(rrQueue.RrdpModelChan),
		" + ", len(rpkiNotifies))
	for _, rpkiNotify := range rpkiNotifies {
		belogs.Debug("addRpkiNotifiesToRpQueue():will PreCheckRrdpUrl, rrQueue.RrdpingParsingCount: ",
			atomic.LoadInt64(&rrQueue.RrdpingParsingCount),
			"   rpkiNotify:", rpkiNotify)
		if !rrQueue.PreCheckRrdpUrl(rpkiNotify) {
			belogs.Debug("addRpkiNotifiesToRpQueue():PreCheckRrdpUrl have exist, before RrdpingParsingCount-1:", rpkiNotify, atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
			atomic.AddInt64(&rrQueue.RrdpingParsingCount, -1)
			belogs.Debug("addRpkiNotifiesToRpQueue():PreCheckRrdpUrl have exist, after RrdpingParsingCount-1:", rpkiNotify, atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
			continue
		}
		belogs.Debug("addRpkiNotifiesToRpQueue():will AddRrdpUrl rpkiNotify: ", rpkiNotify)
		go rrQueue.AddRrdpUrl(rpkiNotify, conf.VariableString("rrdp::destPath")+"/")
	}
}

// will try fail urls  to rrdp again
func tryAgainFailRrdpUrls() bool {
	// try again
	belogs.Debug("TryAgainFailRrdpUrls():try fail urls again: rrQueue.RrdpResult.FailUrls:", jsonutil.MarshalJson(rrQueue.RrdpResult.FailUrls),
		"      rrQueue.RrdpResult.FailUrlsTryCount:", rrQueue.RrdpResult.FailUrlsTryCount)
	if rrQueue.RrdpResult.FailUrlsTryCount <= uint64(conf.Int("rrdp::failRrdpUrlsTryCount")) {
		failRrdpUrls := make([]string, 0)
		//for failRrdpUrl, _ := range
		rrQueue.RrdpResult.FailUrls.Range(func(key, v interface{}) bool {
			failRrdpUrl := key.(string)
			failRrdpUrls = append(failRrdpUrls, failRrdpUrl)
			// delete saved url ,so can try again
			rrQueue.DelRrdpAddedUrl(failRrdpUrl)
			// delete in range, is ok
			rrQueue.RrdpResult.FailUrls.Delete(failRrdpUrl)
			return true
		})
		// clear fail rrdp urls
		rrQueue.RrdpResult.FailUrls = sync.Map{}

		belogs.Debug("TryAgainFailRrdpUrls(): failRysncUrl:", len(failRrdpUrls), failRrdpUrls,
			"   rrQueue.RrdpResult.FailUrlsTryCount: ", rrQueue.RrdpResult.FailUrlsTryCount)
		atomic.AddUint64(&rrQueue.RrdpResult.FailUrlsTryCount, 1)
		belogs.Debug("TryAgainFailRrdpUrls():after  rrQueue.RrdpResult.FailUrlsTryCount: ", rrQueue.RrdpResult.FailUrlsTryCount)

		// check rrdp concurrent count, wait some time,
		rrdpConcurrentCount := conf.Int("rrdp::rrdpConcurrentCount")
		atomic.AddInt64(&rrQueue.RrdpingParsingCount, int64(len(failRrdpUrls)))
		belogs.Debug("TryAgainFailRrdpUrls(): len(failRrdpUrls):", len(failRrdpUrls),
			"  failRrdpUrls:", failRrdpUrls)

		for i, failRrdpUrl := range failRrdpUrls {
			curRrdpingCount := int(atomic.LoadInt64(&rrQueue.CurRrdpingCount))
			if curRrdpingCount <= 2 {

			} else if curRrdpingCount > 2 && curRrdpingCount <= rrdpConcurrentCount {
				belogs.Debug("TryAgainFailRrdpUrls():waitForRrdpUrl, i is smaller, i: ", i,
					" , will wait  1:", 1)
				waitForRrdpUrl(1, failRrdpUrl)
			} else {
				belogs.Debug("TryAgainFailRrdpUrls():waitForRrdpUrl, i is bigger, i: ", i,
					" , will wait  curRrdpingCount+1:", curRrdpingCount+1)
				waitForRrdpUrl(1+curRrdpingCount/2, failRrdpUrl)
			}
			go rrQueue.AddRrdpUrl(failRrdpUrl, conf.VariableString("rrdp::destPath")+"/")
		}
		return true
	}
	return false
}

//rrdp  should wait for some url, because some nic limited access frequency
func waitForRrdpUrl(curRrdpCount int, url string) {

	if curRrdpCount == 0 {
		return
	}
	belogs.Debug("waitForRrdpUrl(): curRrdpCount : ", curRrdpCount, "  will add:", conf.Int("rrdp::rrdpDefaultWaitMs"), " 2* runtime.NumCPU():", 2*runtime.NumCPU())
	curRrdpCount = curRrdpCount + conf.Int("rrdp::rrdpDefaultWaitMs") + 2*runtime.NumCPU()

	// apnic and afrinic should not visit too often
	if strings.Contains(url, "rpki.apnic.net") {
		curRrdpCount = curRrdpCount * 2
	} else if strings.Contains(url, "rpki.afrinic.net") {
		curRrdpCount = curRrdpCount * 10
	}
	min := uint(conf.Int("rrdp::rrdpPerDelayMs") * curRrdpCount)
	randR := uint(conf.Int("rrdp::rrdpDelayRandMs"))
	rand := randutil.IntRange(min, randR)
	belogs.Debug("waitForRrdpUrl():after rand, url is :", url, ",  curRrdpCount is:", curRrdpCount, ", will sleep rand ms:", rand)
	time.Sleep(time.Duration(rand) * time.Millisecond)
}
