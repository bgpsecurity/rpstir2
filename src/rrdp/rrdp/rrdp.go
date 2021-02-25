package rrdp

import (
	"os"
	"sync/atomic"
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	httpclient "github.com/cpusoft/goutil/httpclient"
	jsonutil "github.com/cpusoft/goutil/jsonutil"

	"model"
	"rrdp/db"
	rrdpmodel "rrdp/model"
)

var rrQueue *rrdpmodel.RrdpParseQueue

// start to rrdp
func Start(syncUrls *model.SyncUrls) {
	belogs.Info("Start(): rrdp: syncUrls:", jsonutil.MarshalJson(syncUrls))

	syncRrdpLogs, err := db.GetLastSyncRrdpLogsByNotifyUrl()
	if err != nil {
		belogs.Error("Start(): rrdp: GetLastSyncRrdpLogsByNotifyUrl fail:", err)
		return
	}

	//start rrQueue and rrdpForSelect
	rrQueue = rrdpmodel.NewQueue()
	rrQueue.LastSyncRrdpLogs = syncRrdpLogs
	rrQueue.LabRpkiSyncLogId = syncUrls.SyncLogId
	belogs.Debug("Start(): before startRrdpServer rrQueue:", jsonutil.MarshalJson(rrQueue))

	go startRrdpServer()
	belogs.Debug("Start(): after startRrdpServer rrQueue:", jsonutil.MarshalJson(rrQueue))

	// start to rrdp by sync url in tal, to get root cer
	// first: remove all root cer, so can will rrdp download and will trigger parse all cer files.
	// otherwise, will have to load all root file manually
	os.RemoveAll(conf.VariableString("rrdp::destPath") + "/root/")
	os.MkdirAll(conf.VariableString("rrdp::destPath")+"/root/", os.ModePerm)
	atomic.AddInt64(&rrQueue.RrdpingParsingCount, int64(len(syncUrls.RrdpUrls)))
	belogs.Debug("Start():after RrdpingParsingCount:", atomic.LoadInt64(&rrQueue.RrdpingParsingCount))
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
			belogs.Debug("startRrdpServer():try fail urls again: len(rrQueue.RrdpResult.FailRrdpUrls):", len(rrQueue.RrdpResult.FailUrls))
			if tryAgainFailRrdpUrls() {
				belogs.Debug("startRrdpServer(): tryAgainFailRrdpUrls continue")
				continue
			}
			rrQueue.RrdpResult.EndTime = time.Now()
			rrQueue.RrdpResult.OkUrls = rrQueue.GetRrdpUrls()
			rrQueue.RrdpResult.OkUrlsLen = uint64(len(rrQueue.RrdpResult.OkUrls))
			rrdpResultJson := jsonutil.MarshalJson(rrQueue.RrdpResult)
			belogs.Debug("startRrdpServer():end this rrdp success: rrdpResultJson:", rrdpResultJson)
			// will call sync to return result
			go func(rrdpResultJson string) {
				belogs.Debug("startRrdpServer():call /sync/rrdpresult: rrdpResultJson:", rrdpResultJson)
				httpclient.Post("https://"+conf.String("rpstir2::serverHost")+":"+conf.String("rpstir2::serverHttpsPort")+
					"/sync/rrdpresult", rrdpResultJson, false)
			}(rrdpResultJson)

			// close rrQueue
			rrQueue.Close()

			// return out of the for
			belogs.Info("startRrdpServer():end this rrdp success: rrdpResultJson:", rrdpResultJson, "  time(s):", time.Now().Sub(start).Seconds())
			return
		}
	}
}
