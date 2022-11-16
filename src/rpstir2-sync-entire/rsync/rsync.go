package rsync

import (
	"os"
	"sync/atomic"
	"time"

	model "rpstir2-model"
	"rpstir2-sync-core/rsync"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
)

var rpQueue *RsyncParseQueue

// start to rsync
func rsyncStart(syncUrls *model.SyncUrls) {

	belogs.Info("rsyncStart(): rsync: syncUrls:", jsonutil.MarshalJson(syncUrls))

	//start rpQueue and rsyncForSelect
	rpQueue = NewQueue()
	go startRsyncServer()

	rpQueue.LabRpkiSyncLogId = syncUrls.SyncLogId
	belogs.Debug("rsyncStart(): rpQueue:", jsonutil.MarshalJson(rpQueue))

	// start to rsync by sync url in tal, to get root cer
	// first: remove all root cer, so can will rsync download and will trigger parse all cer files.
	// otherwise, will have to load all root file manually
	os.RemoveAll(conf.VariableString("rsync::destPath") + "/root/")
	os.MkdirAll(conf.VariableString("rsync::destPath")+"/root/", os.ModePerm)
	atomic.AddInt64(&rpQueue.RsyncingParsingCount, int64(len(syncUrls.RsyncUrls)))
	belogs.Debug("rsyncStart():after RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
	for _, url := range syncUrls.RsyncUrls {
		go rpQueue.AddRsyncUrl(url, conf.VariableString("rsync::destPath")+"/root/")
	}

}

// start server ,wait input channel
func startRsyncServer() {
	start := time.Now()
	belogs.Info("startRsyncServer():start")

	for {
		select {
		case rsyncModelChan := <-rpQueue.RsyncModelChan:
			belogs.Debug("startRsyncServer(): rsyncModelChan:", rsyncModelChan,
				"  len(rsyncrpQueue.RsyncModelChan):", len(rpQueue.RsyncModelChan),
				"  receive rsyncModelChan rpQueue.RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
			go rsyncByUrl(rsyncModelChan)
		case parseModelChan := <-rpQueue.ParseModelChan:
			belogs.Debug("startRsyncServer(): parseModelChan:", parseModelChan,
				"  receive parseModelChan rpQueue.RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
			go parseCerFiles(parseModelChan)
		case rsyncParseEndChan := <-rpQueue.RsyncParseEndChan:
			belogs.Debug("startRsyncServer():rsyncParseEndChan:", rsyncParseEndChan, "  rpQueue.RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))

			// try again the fail urls
			//belogs.Debug("startRsyncServer():try fail urls again: len(rpQueue.RsyncResult.FailRsyncUrls):", jsonutil.MarshalJson(rpQueue.RsyncResult.FailUrls))
			//if tryAgainFailRsyncUrls() {
			//	belogs.Debug("startRsyncServer(): tryAgainFailRsyncUrls continue")
			//		continue
			//	}

			// call FoundDiffFiles
			belogs.Debug("startRsyncServer():call FoundDiffFiles")
			var err error
			rpQueue.RsyncResult.AddFilesLen, rpQueue.RsyncResult.DelFilesLen,
				rpQueue.RsyncResult.UpdateFilesLen, rpQueue.RsyncResult.NoChangeFilesLen, err =
				rsync.FoundDiffFiles(rpQueue.LabRpkiSyncLogId, conf.VariableString("rsync::destPath")+"/", nil)
			if err != nil {
				belogs.Error("startRsyncServer(): FoundDiffFiles fail:", err)
				// no return
			}
			rpQueue.RsyncResult.EndTime = time.Now()
			rpQueue.RsyncResult.OkUrls = rpQueue.GetRsyncUrls()
			rpQueue.RsyncResult.OkUrlsLen = uint64(len(rpQueue.RsyncResult.OkUrls))
			rsyncResultJson := jsonutil.MarshalJson(rpQueue.RsyncResult)
			belogs.Debug("startRsyncServer():end this rsync success: rsyncResultJson:", rsyncResultJson)
			// will call sync to return result
			go func(rsyncResultJson string) {
				belogs.Debug("startRsyncServer():call /entiresync/rsyncresult: rsyncResultJson:", rsyncResultJson)
				httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
					"/entiresync/rsyncresult", rsyncResultJson, false)
			}(rsyncResultJson)

			// close rpQueue
			rpQueue.Close()

			// return out of the for
			belogs.Info("startRsyncServer():end this rsync success: rsyncResultJson:", rsyncResultJson, "  time(s):", time.Since(start))
			return
		}
	}
}

func LocalStart(syncUrls *model.SyncUrls) (rsyncResult model.SyncResult, err error) {
	start := time.Now()
	belogs.Info("LocalStart(): rsync: syncUrls:", jsonutil.MarshalJson(syncUrls))

	rsyncResult.AddFilesLen, rsyncResult.DelFilesLen,
		rsyncResult.UpdateFilesLen, rsyncResult.NoChangeFilesLen, err = rsync.FoundDiffFiles(syncUrls.SyncLogId, conf.VariableString("rsync::destPath")+"/", nil)
	if err != nil {
		belogs.Error("LocalStart(): FoundDiffFiles fail:", err)
		// no return
	}
	rsyncResult.EndTime = time.Now()

	belogs.Info("LocalStart():end this rsync success: rsyncResultJson:", jsonutil.MarshalJson(rsyncResult),
		"  time(s):", time.Since(start))
	return rsyncResult, nil

}
