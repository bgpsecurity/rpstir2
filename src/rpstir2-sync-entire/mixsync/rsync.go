package mixsync

import (
	"sync/atomic"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/rsyncutil"
)

func rsyncByUrl(spQueue *SyncParseQueue, syncChan SyncChan) {
	defer func() {
		belogs.Debug("rsyncByUrl():defer spQueue.SyncingAndParsingCount:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		if atomic.LoadInt64(&spQueue.SyncingAndParsingCount) == 0 {
			belogs.Debug("rsyncByUrl(): call SyncAndParseEndChan{}, SyncingAndParsingCount is 0")
			spQueue.SyncAndParseEndChan <- SyncAndParseEndChan{}
		}
	}()

	// start rsync and check err
	// if have error, should set SyncingAndParsingCount -1
	start := time.Now()

	// SyncingCount should +1 and then -1
	atomic.AddInt64(&spQueue.SyncingCount, 1)
	belogs.Debug("rsyncByUrl(): before rsync, syncChan:", syncChan, "    SyncingCount:", atomic.LoadInt64(&spQueue.SyncingCount))
	rsyncutil.SetTimeout(24)
	defer rsyncutil.ResetAllTimeout()
	rsyncDestPath, _, err := rsyncutil.RsyncQuiet(syncChan.Url, syncChan.Dest)
	atomic.AddInt64(&spQueue.SyncingCount, -1)
	belogs.Debug("rsyncByUrl(): rsync syncChan:", syncChan, "     SyncingCount:", atomic.LoadInt64(&spQueue.SyncingCount),
		"     rsyncDestPath:", rsyncDestPath)
	if err != nil {
		spQueue.SyncResult.FailUrls.Store(syncChan.Url, err.Error())
		belogs.Error("rsyncByUrl():RsyncQuiet fail, syncChan.Url:", syncChan.Url, "   err:", err, "  time(s):", time.Since(start))
		belogs.Debug("rsyncByUrl():RsyncQuiet fail, before SyncingAndParsingCount-1:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		atomic.AddInt64(&spQueue.SyncingAndParsingCount, -1)
		belogs.Debug("rsyncByUrl():RsyncQuiet fail, after SyncingAndParsingCount-1:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		return
	}
	belogs.Debug("rsyncByUrl(): rsync.Rsync url:", syncChan.Url, "   rsyncDestPath:", rsyncDestPath)

	filePathNames := make([]string, 0)
	filePathNames = append(filePathNames, rsyncDestPath)
	parseChan := ParseChan{Url: syncChan.Url, FilePathNames: filePathNames}
	belogs.Debug("rsyncByUrl():before parseChan:", jsonutil.MarshalJson(parseChan), "   len(spQueue.ParseChan):", len(spQueue.ParseChan))
	spQueue.ParseChan <- parseChan
	belogs.Info("rsyncByUrl(): after parseChan:", jsonutil.MarshalJson(parseChan),
		"     SyncingCount:", atomic.LoadInt64(&spQueue.SyncingCount),
		"     rsyncDestPath:", rsyncDestPath,
		"     time(s):", time.Since(start))

}

func parseRsyncCerFiles(spQueue *SyncParseQueue, parseChan ParseChan) {
	defer func() {
		belogs.Debug("parseRsyncCerFiles():defer spQueue.SyncingAndParsingCount:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		if atomic.LoadInt64(&spQueue.SyncingAndParsingCount) == 0 {
			belogs.Debug("parseRsyncCerFiles(): call SyncAndParseEndChan{}, SyncingAndParsingCount is 0")
			spQueue.SyncAndParseEndChan <- SyncAndParseEndChan{}
		}
	}()
	belogs.Debug("parseRsyncCerFiles(): parseChan:", jsonutil.MarshalJson(parseChan))
	if len(parseChan.FilePathNames) == 0 {
		return
	}

	// if have erorr, should set SyncingAndParsingCount -1
	// get all cer files, include subcer
	m := make(map[string]string, 0)
	m[".cer"] = ".cer"
	cerFiles, err := osutil.GetAllFilesBySuffixs(parseChan.FilePathNames[0], m)
	if err != nil {
		belogs.Error("parseRsyncCerFiles():GetAllFilesBySuffixs fail, parseChan.FilePathNames[0]:", parseChan.FilePathNames[0], "   err:", err)
		belogs.Debug("parseRsyncCerFiles():GetAllFilesBySuffixs, before SyncingAndParsingCount-1:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		atomic.AddInt64(&spQueue.SyncingAndParsingCount, -1)
		belogs.Debug("parseRsyncCerFiles():GetAllFilesBySuffixs, after SyncingAndParsingCount-1:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		return
	}
	belogs.Debug("parseRsyncCerFiles(): len(cerFiles):", len(cerFiles))

	// if there are no cer files, return
	if len(cerFiles) == 0 {
		belogs.Debug("parseRsyncCerFiles():len(cerFiles)==0, before SyncingAndParsingCount-1:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		atomic.AddInt64(&spQueue.SyncingAndParsingCount, -1)
		belogs.Debug("parseRsyncCerFiles():len(cerFiles)==0, after SyncingAndParsingCount-1:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
		return
	}
	belogs.Debug("parseRsyncCerFiles(): parseChan.Url:", parseChan.Url, "  cerFiles:", cerFiles)
	belogs.Info("parseRrdpCerFiles():  parseChan.Url:", parseChan.Url, "  len(cerFiles):", len(cerFiles))
	parseCerAndGetSubRepoUrlAndAddToSpQueue(spQueue, cerFiles)
	belogs.Debug("parseRsyncCerFiles(): after parseCerAndGetSubRepoUrlAndAddToSpQueue cerFiles:", cerFiles)
}
