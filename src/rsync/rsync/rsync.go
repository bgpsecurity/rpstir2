package rsync

import (
	"os"
	"sync/atomic"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	httpclient "github.com/cpusoft/goutil/httpclient"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	rsyncmodel "rsync/model"
)

var rpQueue *rsyncmodel.RsyncParseQueue

// start to rsync
func Start(syncUrls *model.SyncUrls) {

	belogs.Info("Start(): rsync: syncUrls:", jsonutil.MarshalJson(syncUrls))

	//start rpQueue and rsyncForSelect
	rpQueue = rsyncmodel.NewQueue()
	go startRsyncServer()

	rpQueue.LabRpkiSyncLogId = syncUrls.SyncLogId
	belogs.Debug("Start(): rpQueue:", jsonutil.MarshalJson(rpQueue))

	// start to rsync by sync url in tal, to get root cer
	// first: remove all root cer, so can will rsync download and will trigger parse all cer files.
	// otherwise, will have to load all root file manually
	os.RemoveAll(conf.VariableString("rsync::destPath") + "/root/")
	os.MkdirAll(conf.VariableString("rsync::destPath")+"/root/", os.ModePerm)
	atomic.AddInt64(&rpQueue.RsyncingParsingCount, int64(len(syncUrls.RsyncUrls)))
	belogs.Debug("Start():after RsyncingParsingCount:", atomic.LoadInt64(&rpQueue.RsyncingParsingCount))
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
			belogs.Debug("startRsyncServer():try fail urls again: len(rpQueue.RsyncResult.FailRsyncUrls):", len(rpQueue.RsyncResult.FailUrls))
			if tryAgainFailRsyncUrls() {
				belogs.Debug("startRsyncServer(): tryAgainFailRsyncUrls continue")
				continue
			}

			// call FoundDiffFiles
			belogs.Debug("startRsyncServer():call FoundDiffFiles")
			var err error
			rpQueue.RsyncResult.AddFilesLen, rpQueue.RsyncResult.DelFilesLen,
				rpQueue.RsyncResult.UpdateFilesLen, rpQueue.RsyncResult.NoChangeFilesLen, err = FoundDiffFiles(rpQueue.LabRpkiSyncLogId)
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
				belogs.Debug("startRsyncServer():call /sync/rsyncresult: rsyncResultJson:", rsyncResultJson)
				httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
					"/sync/rsyncresult", rsyncResultJson, false)
			}(rsyncResultJson)

			// close rpQueue
			rpQueue.Close()

			// return out of the for
			belogs.Info("startRsyncServer():end this rsync success: rsyncResultJson:", rsyncResultJson, "  time(s):", time.Now().Sub(start).Seconds())
			return
		}
	}
}

func LocalStart(syncUrls *model.SyncUrls) (rsyncResult model.SyncResult, err error) {
	start := time.Now()
	belogs.Info("LocalStart(): rsync: syncUrls:", jsonutil.MarshalJson(syncUrls))

	rsyncResult.AddFilesLen, rsyncResult.DelFilesLen,
		rsyncResult.UpdateFilesLen, rsyncResult.NoChangeFilesLen, err = FoundDiffFiles(syncUrls.SyncLogId)
	if err != nil {
		belogs.Error("LocalStart(): FoundDiffFiles fail:", err)
		// no return
	}
	rsyncResult.EndTime = time.Now()

	belogs.Info("LocalStart():end this rsync success: rsyncResultJson:", jsonutil.MarshalJson(rsyncResult),
		"  time(s):", time.Now().Sub(start).Seconds())
	return rsyncResult, nil

}
