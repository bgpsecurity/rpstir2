package rrdp

import (
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	httpclient "github.com/cpusoft/goutil/httpclient"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"
	randutil "github.com/cpusoft/goutil/randutil"
	rrdputil "github.com/cpusoft/goutil/rrdputil"

	"model"
	rrdpmodel "rrdp/model"
)

// needRrdp: have new serial ,need rrdp to sync
func rrdpByUrlImpl(notifyUrl string, destPath string, syncLogId uint64,
	lastSyncRrdpLogs map[string]model.LabRpkiSyncRrdpLog) (needRrdp bool, rrdpFiles []rrdputil.RrdpFile, err error) {
	start := time.Now()
	//defer RrdpByUrlDefer(rrdpUrl, rrdpUrlCh)
	belogs.Debug("rrdpByUrlImpl():start, notifyUrl, destPath, syncLogId:", notifyUrl, destPath, syncLogId)

	// get notify xml
	notificationModel, err := getRrdpNotification(notifyUrl)
	if err != nil {
		belogs.Error("rrdpByUrlImpl(): GetRrdpNotification fail, notifyUrl, syncLogId: ",
			notifyUrl, syncLogId, err)
		return false, nil, err
	}

	lastSyncRrdpLog, has := lastSyncRrdpLogs[notifyUrl]
	belogs.Debug("rrdpByUrlImpl(): compare :",
		"has, lastSyncRrdpLog.SessionId, notificationModel.SessionId, lastSyncRrdpLog.curSerial, notificationModel.MaxSerail: ",
		has, lastSyncRrdpLog.SessionId, notificationModel.SessionId, lastSyncRrdpLog.CurSerial, notificationModel.MaxSerail)
	if has && lastSyncRrdpLog.SessionId == notificationModel.SessionId && lastSyncRrdpLog.CurSerial == notificationModel.MaxSerail {
		belogs.Debug("rrdpByUrlImpl(): no new rrdp serial, no need rrdp to download, just return:",
			"has:", has,
			",  lastSyncRrdpLog.SessionId, notificationModel.SessionId:", lastSyncRrdpLog.SessionId, notificationModel.SessionId,
			",  lastSyncRrdpLog.CurSerial, notificationModel.MaxSerail:", lastSyncRrdpLog.CurSerial, notificationModel.MaxSerail)
		belogs.Info("rrdpByUrlImpl(): no new rrdp serial, no need rrdp to download, just return:", notifyUrl)
		return false, nil, nil
	}
	canDelta := false
	if has &&
		lastSyncRrdpLog.SessionId == notificationModel.SessionId &&
		lastSyncRrdpLog.CurSerial >= notificationModel.MinSerail &&
		lastSyncRrdpLog.CurSerial < notificationModel.MaxSerail {
		canDelta = true
	}
	belogs.Info("rrdpByUrlImpl():notifyUrl canDelta:", notifyUrl, canDelta)

	// need to get snapshot
	var snapshotDeltaResult rrdpmodel.SnapshotDeltaResult
	if !canDelta {
		snapshotDeltaResult = rrdpmodel.SnapshotDeltaResult{
			NotifyUrl:  notifyUrl,
			DestPath:   destPath,
			LastSerial: 0}
		belogs.Info("rrdpByUrlImpl(): will snapshot:", notifyUrl, snapshotDeltaResult)
		err = processRrdpSnapshot(syncLogId, &notificationModel, &snapshotDeltaResult)

	} else {
		// get delta
		snapshotDeltaResult = rrdpmodel.SnapshotDeltaResult{
			NotifyUrl:  notifyUrl,
			DestPath:   destPath,
			LastSerial: lastSyncRrdpLog.CurSerial}
		belogs.Info("rrdpByUrlImpl(): will delta:", notifyUrl, snapshotDeltaResult)
		err = processRrdpDelta(syncLogId, &notificationModel, &snapshotDeltaResult)
	}

	if err != nil {
		belogs.Error("rrdpByUrlImpl(): GetLastSyncRrdpLog fail, notifyUrl, syncLogIdsyncLogId: ",
			notifyUrl, syncLogId, err)
		return false, nil, err
	}
	belogs.Info("rrdpByUrlImpl(): end ok, notifyUrl, len(files):", notifyUrl, len(snapshotDeltaResult.RrdpFiles), "  time(s):", time.Now().Sub(start).Seconds())
	return true, snapshotDeltaResult.RrdpFiles, nil

}

func rrdpByUrl(rrdpModelChan rrdpmodel.RrdpModelChan) {
	defer func() {
		belogs.Debug("RrdpByUrl():defer rrQueue.RrdpingParsingCount:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount),
			"  rrdpurls:", rrQueue.GetRrdpUrls())
		if atomic.LoadInt64(&rrQueue.RrdpingParsingCount) == 0 {
			belogs.Debug("RrdpByUrl(): call rrdpmodel.RrdpParseEndChan{}, RrdpingParsingCount is 0")
			rrQueue.RrdpParseEndChan <- rrdpmodel.RrdpParseEndChan{}
		}
	}()

	// start rrdp and check err
	// if have error, should set RrdpingParsingCount -1
	start := time.Now()

	// CurRrdpingCount should +1 and then -1
	atomic.AddInt64(&rrQueue.CurRrdpingCount, 1)
	belogs.Debug("RrdpByUrl(): before rrdp, rrdpModelChan:", rrdpModelChan,
		"    CurRrdpingCount:", atomic.LoadInt64(&rrQueue.CurRrdpingCount), "   startTime:", start)

	needRrdp, rrdpFiles, err := rrdpByUrlImpl(rrdpModelChan.Url, rrdpModelChan.Dest, rrQueue.LabRpkiSyncLogId, rrQueue.LastSyncRrdpLogs)
	atomic.AddInt64(&rrQueue.CurRrdpingCount, -1)
	belogs.Debug("RrdpByUrl(): rrdpByUrlImpl, needRrdp,  len(rrdpFiles), err:", needRrdp, len(rrdpFiles), err)

	if err != nil {
		rrQueue.RrdpResult.FailUrls[rrdpModelChan.Url] = err.Error()
		belogs.Error("RrdpByUrl():RrdpQuiet fail, rrdpModelChan.Url:", rrdpModelChan.Url, "   err:", err, "  time(s):", time.Now().Sub(start).Seconds())
		belogs.Debug("RrdpByUrl():RrdpQuiet fail, before RrdpingParsingCount-1:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
		atomic.AddInt64(&rrQueue.RrdpingParsingCount, -1)
		belogs.Debug("RrdpByUrl():RrdpQuiet fail, after RrdpingParsingCount-1:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
		return
	}
	if !needRrdp {
		belogs.Debug("RrdpByUrl():no need rrdp:", rrdpModelChan.Url, "   , before RrdpingParsingCount-1:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
		atomic.AddInt64(&rrQueue.RrdpingParsingCount, -1)
		belogs.Debug("RrdpByUrl():no need rrdp:", rrdpModelChan.Url, "   , after RrdpingParsingCount-1:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
		return
	}
	belogs.Debug("RrdpByUrl(): rrdp.Rrdp url:", rrdpModelChan.Url, "   len(rrdpFiles):", len(rrdpFiles))

	filePathNames := make([]string, 0, len(rrdpFiles))
	for i := range rrdpFiles {
		if osutil.ExtNoDot(rrdpFiles[i].FileName) == "cer" && rrdpFiles[i].SyncType == "add" {
			filePathName := osutil.JoinPathFile(rrdpFiles[i].FilePath, rrdpFiles[i].FileName)
			filePathNames = append(filePathNames, filePathName)
		}
		if rrdpFiles[i].SyncType == "add" {
			atomic.AddUint64(&rrQueue.RrdpResult.AddFilesLen, 1)
		} else if rrdpFiles[i].SyncType == "del" {
			atomic.AddUint64(&rrQueue.RrdpResult.DelFilesLen, 1)
		}
	}
	parseModelChan := rrdpmodel.ParseModelChan{FilePathNames: filePathNames}
	belogs.Debug("RrdpByUrl():parseModelChan:", parseModelChan, "   len(rrQueue.ParseModelChan):", len(rrQueue.ParseModelChan))
	belogs.Info("RrdpByUrl():  CurRrdpingCount:", atomic.LoadInt64(&rrQueue.CurRrdpingCount),
		"  time(s):", time.Now().Sub(start).Seconds())

	rrQueue.ParseModelChan <- parseModelChan

}

func parseCerFiles(parseModelChan rrdpmodel.ParseModelChan) {
	defer func() {
		belogs.Debug("parseCerFiles():defer rrQueue.RrdpingParsingCount:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount),
			"  rrdpurls:", rrQueue.GetRrdpUrls())
		if atomic.LoadInt64(&rrQueue.RrdpingParsingCount) == 0 {
			belogs.Debug("parseCerFiles(): call rrdpmodel.RyncParseEndChan{}, RrdpingParsingCount is 0")
			rrQueue.RrdpParseEndChan <- rrdpmodel.RrdpParseEndChan{}
		}
	}()
	belogs.Debug("parseCerFiles(): parseModelChan:", parseModelChan)

	// if have erorr, should set RrdpingParsingCount -1
	// get all cer files, include subcer

	// foreach every cerfiles to parseCerAndGetSubCaRepositoryUrl
	rpkiNotifies := make([]string, 0)
	for i := range parseModelChan.FilePathNames {
		// just trigger sync ,no need save to db
		rpkiNotify := parseCerAndGetRpkiNotify(parseModelChan.FilePathNames[i])
		if len(rpkiNotify) > 0 {
			rpkiNotifies = append(rpkiNotifies, rpkiNotify)
		}
	}
	belogs.Debug("parseCerFiles(): FilePathNames, rpkiNotifies:", parseModelChan.FilePathNames, rpkiNotifies)

	// check rrdp concurrent count, wait some time,
	// the father rrdpingparsingcount -1 ,and the children rrdpingparsingcount + len()
	belogs.Debug("parseCerFiles():will add rpkiNotifies, before RrdpingParsingCount:",
		atomic.LoadInt64(&rrQueue.RrdpingParsingCount), "  len(rpkiNotifies)-1:", len(rpkiNotifies)-1)
	atomic.AddInt64(&rrQueue.RrdpingParsingCount, int64(len(rpkiNotifies)-1))
	belogs.Debug("parseCerFiles():will add rpkiNotifies, after RrdpingParsingCount:",
		atomic.LoadInt64(&rrQueue.RrdpingParsingCount), "  len(rpkiNotifies)-1:", len(rpkiNotifies)-1)
	addRpkiNotifiesToRpQueue(rpkiNotifies)
}

// call /parsevalidate/parse to parse cert, and save result
func parseCerAndGetRpkiNotify(cerFile string) (RpkiNotify string) {

	// call parse, not need to save body to db
	start := time.Now()
	belogs.Debug("parseCerAndGetRpkiNotify():/parsevalidate/parsefilesimple cerFile:", cerFile)
	// post file, still use http
	resp, body, err := httpclient.PostFile("http", conf.String("rpstir2::serverHost"), conf.Int("rpstir2::serverHttpPort"),
		"/parsevalidate/parsefilesimple", cerFile, "")
	belogs.Debug("parseCerAndGetRpkiNotify():after /parsevalidate/parsefilesimple cerFile:", cerFile, len(body))

	if err != nil {
		rrQueue.RrdpResult.FailParseValidateCerts[cerFile] = err.Error()
		belogs.Error("parseCerAndGetRpkiNotify(): filerepo file connect failed:", cerFile, "   err:", err)
		return ""
	}
	defer resp.Body.Close()

	// get parse result
	parseCerSimpleResponse := model.ParseCerSimpleResponse{}
	jsonutil.UnmarshalJson(string(body), &parseCerSimpleResponse)
	belogs.Debug("parseCerAndGetRpkiNotify(): get from parsecert, parseCerSimpleResponse.Result:", parseCerSimpleResponse.Result)
	if parseCerSimpleResponse.HttpResponse.Result != "ok" {
		belogs.Error("parseCerAndGetRpkiNotify(): parsecert file failed:", cerFile, "   err:", parseCerSimpleResponse.HttpResponse.Msg)
		rrQueue.RrdpResult.FailParseValidateCerts[cerFile] = parseCerSimpleResponse.HttpResponse.Msg
		return ""
	}

	// get the sub repo url in cer, and send it to rpqueue
	belogs.Info("parseCerAndGetRpkiNotify(): cerFile:", cerFile, "    caRepository:", parseCerSimpleResponse.ParseCerSimple.CaRepository,
		"  time(s):", time.Now().Sub(start).Seconds())
	return parseCerSimpleResponse.ParseCerSimple.RpkiNotify

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
	belogs.Debug("TryAgainFailRrdpUrls():try fail urls again: len(rrQueue.RrdpResult.FailUrls):", len(rrQueue.RrdpResult.FailUrls),
		"      rrQueue.RrdpResult.FailUrlsTryCount:", rrQueue.RrdpResult.FailUrlsTryCount)
	if len(rrQueue.RrdpResult.FailUrls) > 0 &&
		rrQueue.RrdpResult.FailUrlsTryCount <= uint64(conf.Int("rrdp::failRrdpUrlsTryCount")) {
		failRrdpUrls := make([]string, 0, len(rrQueue.RrdpResult.FailUrls))
		for failRrdpUrl := range rrQueue.RrdpResult.FailUrls {
			failRrdpUrls = append(failRrdpUrls, failRrdpUrl)
			// delete saved url ,so can try again
			rrQueue.DelRrdpAddedUrl(failRrdpUrl)
		}
		// clear fail rrdp urls
		rrQueue.RrdpResult.FailUrls = make(map[string]string, 200)

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
