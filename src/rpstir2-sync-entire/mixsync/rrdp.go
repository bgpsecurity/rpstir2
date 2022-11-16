package mixsync

import (
	"sync/atomic"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"rpstir2-sync-core/rrdp"
)

func rrdpByUrl(spQueue *SyncParseQueue, syncChan SyncChan) {
	defer func() {
		belogs.Debug("rrdpByUrl():defer spQueue.SyncingAndParsingCount:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount),
			"  rrdpurls:", spQueue.GetSyncUrls())
		if atomic.LoadInt64(&spQueue.SyncingAndParsingCount) == 0 {
			belogs.Info("rrdpByUrl(): call SyncAndParseEndChan{}, SyncingAndParsingCount is 0")
			spQueue.SyncAndParseEndChan <- SyncAndParseEndChan{}
		}
	}()

	// start rrdp and check err
	// if have error, should set SyncingAndParsingCount -1
	start := time.Now()

	belogs.Debug("rrdpByUrl(): before rrdp, syncChan:", jsonutil.MarshalJson(syncChan),
		"    SyncingCount:", atomic.LoadInt64(&spQueue.SyncingCount), "   startTime:", start)
	lastSyncRrdpLog, hasLast := spQueue.LastSyncRrdpLogs[syncChan.Url]
	rrdpByUrlModel := rrdp.RrdpByUrlModel{
		NotifyUrl:     syncChan.Url,
		DestPath:      syncChan.Dest,
		HasPath:       true,
		HasLast:       hasLast,
		LastSessionId: lastSyncRrdpLog.SessionId,
		LastCurSerial: lastSyncRrdpLog.CurSerial,
		SyncLogId:     spQueue.LabRpkiSyncLogId,
	}
	belogs.Debug("rrdpByUrl():rrdpByUrlModel:", jsonutil.MarshalJson(rrdpByUrlModel))
	// will ignore connectRrdpUrlCh
	connectRrdpUrlCh := make(chan bool, 1)

	// SyncingCount should +1 and then -1
	atomic.AddInt64(&spQueue.SyncingCount, 1)
	rrdpFiles, err := rrdp.RrdpByUrlImpl(rrdpByUrlModel, connectRrdpUrlCh, nil)
	atomic.AddInt64(&spQueue.SyncingCount, -1)
	belogs.Debug("rrdpByUrl(): RrdpByUrlImpl, len(rrdpFiles), err:", len(rrdpFiles), err)

	if err != nil {
		spQueue.SyncResult.FailUrls.Store(syncChan.Url, err.Error())
		belogs.Error("rrdpByUrl():RrdpByUrlImpl fail, syncChan.Url:", syncChan.Url, "   err:", err, "  time(s):", time.Since(start))
		belogs.Debug("rrdpByUrl():RrdpByUrlImpl fail, before SyncingAndParsingCount-1:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		atomic.AddInt64(&spQueue.SyncingAndParsingCount, -1)
		belogs.Debug("rrdpByUrl():RrdpByUrlImpl fail, after SyncingAndParsingCount-1:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		return
	}
	if len(rrdpFiles) == 0 {
		belogs.Debug("rrdpByUrl():len(rrdpFiles) == 0,no need rrdp,before:", syncChan.Url, "   , before SyncingAndParsingCount-1:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		atomic.AddInt64(&spQueue.SyncingAndParsingCount, -1)
		belogs.Debug("rrdpByUrl():len(rrdpFiles) == 0, no need rrdp,after:", syncChan.Url, "   , after SyncingAndParsingCount-1:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		return
	}
	belogs.Info("rrdpByUrl(): syncChan.Url:", syncChan.Url, "   len(rrdpFiles):", len(rrdpFiles))

	filePathNames := make([]string, 0)
	for i := range rrdpFiles {
		if osutil.ExtNoDot(rrdpFiles[i].FileName) == "cer" &&
			(rrdpFiles[i].SyncType == "add" || rrdpFiles[i].SyncType == "update") {
			filePathName := osutil.JoinPathFile(rrdpFiles[i].FilePath, rrdpFiles[i].FileName)
			filePathNames = append(filePathNames, filePathName)
		}
		if rrdpFiles[i].SyncType == "add" {
			atomic.AddUint64(&spQueue.SyncResult.AddFilesLen, 1)
		} else if rrdpFiles[i].SyncType == "del" {
			atomic.AddUint64(&spQueue.SyncResult.DelFilesLen, 1)
		} else if rrdpFiles[i].SyncType == "update" {
			atomic.AddUint64(&spQueue.SyncResult.UpdateFilesLen, 1)
		}

		belogs.Debug("rrdpByUrl(): rrdpFiles[i]:", jsonutil.MarshalJson(rrdpFiles[i]))
	}
	parseChan := ParseChan{Url: syncChan.Url, FilePathNames: filePathNames}
	belogs.Debug("rrdpByUrl(): before parseChan:", jsonutil.MarshalJson(parseChan), "   len(spQueue.ParseChan):", len(spQueue.ParseChan))
	spQueue.ParseChan <- parseChan
	belogs.Info("rrdpByUrl(): after parseChan:", jsonutil.MarshalJson(parseChan),
		"    SyncingCount:", atomic.LoadInt64(&spQueue.SyncingCount),
		"    len(filePathNames):", len(filePathNames),
		"    len(rrdpFiles):", len(rrdpFiles),
		"    time(s):", time.Since(start))

}

func parseRrdpCerFiles(spQueue *SyncParseQueue, parseChan ParseChan) {
	defer func() {
		belogs.Debug("parseRrdpCerFiles():defer spQueue.SyncingAndParsingCount:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount),
			"  rrdpurls:", spQueue.GetSyncUrls())
		if atomic.LoadInt64(&spQueue.SyncingAndParsingCount) == 0 {
			belogs.Info("parseRrdpCerFiles(): call SyncAndParseEndChan{}, SyncingAndParsingCount is 0")
			spQueue.SyncAndParseEndChan <- SyncAndParseEndChan{}
		}
	}()
	belogs.Debug("parseRrdpCerFiles(): parseChan:", jsonutil.MarshalJson(parseChan))
	belogs.Info("parseRrdpCerFiles():  parseChan.Url:", parseChan.Url, "   len(parseChan.FilePathNames):", len(parseChan.FilePathNames))
	parseCerAndGetSubRepoUrlAndAddToSpQueue(spQueue, parseChan.FilePathNames)
	belogs.Debug("parseRrdpCerFiles(): after parseCerAndGetSubRepoUrlAndAddToSpQueue parseChan:", jsonutil.MarshalJson(parseChan))
}
