package mixsync

import (
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/urlutil"
	"rpstir2-sync-core/rsync"
	coresync "rpstir2-sync-core/sync"
)

// start server ,wait input channel
func startSyncServer(spQueue *SyncParseQueue, syncState *SyncState, syncServerWg *sync.WaitGroup) {
	defer func() {
		belogs.Info("startSyncServer():syncServerWg.Done()")
		syncServerWg.Done()
	}()

	start := time.Now()
	belogs.Info("startSyncServer():start")

	for {
		select {
		case syncChan := <-spQueue.SyncChan:
			belogs.Debug("startSyncServer():receive syncChan:", jsonutil.MarshalJson(syncChan),
				"  len(spQueue.SyncChan):", len(spQueue.SyncChan),
				"  receive syncChan spQueue.SyncingAndParsingCount:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
			go syncByUrl(spQueue, syncChan)
		case parseChan := <-spQueue.ParseChan:
			belogs.Debug("startSyncServer():receive parseChan:", jsonutil.MarshalJson(parseChan),
				"  receive parseChan spQueue.SyncingAndParsingCount:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
			go parseCerFiles(spQueue, parseChan)
		case syncAndParseEndChan := <-spQueue.SyncAndParseEndChan:
			belogs.Debug("startSyncServer():receive syncAndParseEndChan:", syncAndParseEndChan, "  spQueue.SyncingAndParsingCount:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))

			// get rsync diff files
			var err error
			addFilesLen, delFilesLen, updateFilesLen, noChangeFilesLen, err :=
				rsync.FoundDiffFiles(spQueue.LabRpkiSyncLogId, conf.VariableString("rsync::destPath")+"/", nil)
			if err != nil {
				belogs.Error("startSyncServer(): FoundDiffFiles fail:", err)
				// no return
			}
			belogs.Debug("startSyncServer():call FoundDiffFiles, addFilesLen, delFilesLen, updateFilesLen, noChangeFilesLen:",
				addFilesLen, delFilesLen, updateFilesLen, noChangeFilesLen)

			spQueue.SyncResult.AddFilesLen += addFilesLen
			spQueue.SyncResult.DelFilesLen += delFilesLen
			spQueue.SyncResult.UpdateFilesLen += updateFilesLen
			spQueue.SyncResult.NoChangeFilesLen += noChangeFilesLen
			spQueue.SyncResult.EndTime = time.Now()
			spQueue.SyncResult.OkUrls = spQueue.GetSyncUrls()
			spQueue.SyncResult.OkUrlsLen = uint64(len(spQueue.SyncResult.OkUrls))
			syncResultJson := jsonutil.MarshalJson(spQueue.SyncResult)
			belogs.Debug("startSyncServer():syncResultJson:", syncResultJson)

			syncState.EndTime = spQueue.SyncResult.EndTime
			syncState.SyncUrls = spQueue.GetSyncUrls()
			syncState.SyncResult = spQueue.SyncResult
			belogs.Debug("startSyncServer():syncState:", jsonutil.MarshalJson(syncState))

			// close spQueue
			if !spQueue.IsClose() {
				spQueue.Close()
			}

			// return out of the for
			belogs.Info("startSyncServer():end this sync server ssuccess: syncResultJson:", syncResultJson,
				"  time(s):", time.Now().Sub(start))
			return
		}
	}

}

func syncByUrl(spQueue *SyncParseQueue, syncChan SyncChan) {
	belogs.Debug("syncByUrl():syncChan.Url:", syncChan.Url)
	if strings.Contains(syncChan.Url, "https://") {
		err := checkAndDeleteLastSync(true, syncChan.Url)
		if err != nil {
			belogs.Error("syncByUrl():checkAndDeleteLastSync rrdp fail:", syncChan.Url, err)
			return
		}
		rrdpByUrl(spQueue, syncChan)
	} else if strings.Contains(syncChan.Url, "rsync://") {
		err := checkAndDeleteLastSync(false, syncChan.Url)
		if err != nil {
			belogs.Error("syncByUrl():checkAndDeleteLastSync rsync fail:", syncChan.Url, err)
			return
		}
		rsyncByUrl(spQueue, syncChan)
	}
}

func parseCerFiles(spQueue *SyncParseQueue, parseChan ParseChan) {
	belogs.Debug("parseCerFiles():parseChan:", jsonutil.MarshalJson(parseChan))
	if strings.Contains(parseChan.Url, "https://") {
		parseRrdpCerFiles(spQueue, parseChan)
	} else if strings.Contains(parseChan.Url, "rsync://") {
		parseRsyncCerFiles(spQueue, parseChan)
	}
}

func checkAndDeleteLastSync(isRrdp bool, url string) (err error) {
	rrdpDestPath := conf.VariableString("rrdp::destPath") + osutil.GetPathSeparator()
	rsyncDestPath := conf.VariableString("rsync::destPath") + osutil.GetPathSeparator()
	belogs.Debug("checkAndDeleteLastSync(): rrdpDestPath:", rrdpDestPath, "   rsyncDestPath:", rsyncDestPath,
		"  isRrdp:", isRrdp, "     url:", url)
	// last sync, may be another sync type
	testLastSyncDestPath := ""
	if isRrdp {
		testLastSyncDestPath, err = urlutil.JoinPrefixPathAndUrlHost(rsyncDestPath, url)
	} else {
		testLastSyncDestPath, err = urlutil.JoinPrefixPathAndUrlHost(rrdpDestPath, url)
	}
	if err != nil {
		belogs.Error("checkAndDeleteLastSync(): JoinPrefixPathAndUrlHost fail, testLastSyncDestPath:", testLastSyncDestPath, err)
		return err
	}
	// check if exists in local directory
	exists, err := osutil.IsExists(testLastSyncDestPath)
	if err != nil {
		belogs.Error("checkAndDeleteLastSync(): IsExists fail ,testLastSyncDestPath:", testLastSyncDestPath, err)
		return err
	}
	if !exists {
		belogs.Debug("checkAndDeleteLastSync(): url no exists in another destPath from last sync:",
			" isRrdp:", isRrdp, "   url:", url, "  testLastSyncDestPath:", testLastSyncDestPath)
		return nil
	}

	// delete last record in mysql
	belogs.Debug("checkAndDeleteLastSync(): url exists in another destPath from last sync, need delete all last sync record in db for this url:",
		" isRrdp:", isRrdp, "   url:", url, "  testLastSyncDestPath:", testLastSyncDestPath)
	err = coresync.DelByFilePathDb(testLastSyncDestPath)
	if err != nil {
		belogs.Error("checkAndDeleteLastSync():DelByFilePathDb fail, testLastSyncDestPath:", testLastSyncDestPath, err)
		return err
	}

	// delete in disk
	err = os.RemoveAll(testLastSyncDestPath)
	if err != nil {
		belogs.Error("checkAndDeleteLastSync():RemoveAll fail, testLastSyncDestPath:", testLastSyncDestPath, err)
		// not return err
	}

	belogs.Info("checkAndDeleteLastSync(): delete all last sync record in db for this url:",
		" isRrdp:", isRrdp, "   url:", url, " will del testLastSyncDestPath:", testLastSyncDestPath)
	return nil

}
