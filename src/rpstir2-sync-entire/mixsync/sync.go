package mixsync

import (
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	model "rpstir2-model"
	"rpstir2-sync-core/rrdp"
	coresync "rpstir2-sync-core/sync"
)

func syncStart(syncStyle model.SyncStyle) (nextStep string, err error) {
	start := time.Now()

	belogs.Info("syncStart():syncStyle:", syncStyle)

	syncState := SyncState{StartTime: time.Now(), SyncStyle: syncStyle.SyncStyle}

	// syncStyle: sync/rsync/rrdp,state: syncing;
	syncLogId, err := coresync.InsertSyncLogStartDb(syncStyle.SyncStyle, "syncing")
	if err != nil {
		belogs.Error("syncStart():InsertSyncLogStartDb fail:", err)
		return "", err
	}
	belogs.Info("syncStart():syncLogId:", syncLogId, "  syncState:", jsonutil.MarshalJson(syncState))

	// call tals , get all tals
	talModels, err := getTals()
	if err != nil {
		belogs.Error("syncStart(): getTals failed, err:", err)
		return "", err
	}
	belogs.Debug("syncStart(): len(talModels):", len(talModels))

	// call rrdp and rsync and wait for result
	err = callSync(syncLogId, talModels, &syncState)
	if err != nil {
		belogs.Error("syncStart():callSync fail:", err)
		return "", err
	}

	// update lab_rpki_sync_log
	err = coresync.UpdateSyncLogEndDb(syncLogId, "synced", jsonutil.MarshalJson(syncState))
	if err != nil {
		belogs.Error("syncStart():UpdateSyncLogEndDb fail:", err)
		return "", err
	}
	belogs.Info("syncStart(): end sync, will parsevalidate,  time(s):", time.Since(start))

	return "parsevalidate", nil

}

func getTals() (talModels []model.TalModel, err error) {
	start := time.Now()
	// by /tal/gettals
	talModelsResponse := model.TalModelsResponse{}
	err = httpclient.PostAndUnmarshalResponseModel("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
		"/tal/gettals", "", false, &talModelsResponse)
	if err != nil {
		belogs.Error("getTals(): /tal/gettals failed, err:", err)
		return nil, err
	}

	belogs.Debug("getTals(): talModelsResponse:",
		jsonutil.MarshalJson(talModelsResponse), "  time(s):", time.Since(start))

	if len(talModelsResponse.TalModels) == 0 {
		belogs.Error("getTals(): there is no tal file")
		return nil, errors.New("there is no tal file")
	}
	return talModelsResponse.TalModels, nil
}

func callSync(syncLogId uint64, talModels []model.TalModel, syncState *SyncState) (err error) {
	start := time.Now()
	belogs.Info("callSync(): syncLogId:", syncLogId, "   talModels:", jsonutil.MarshalJson(talModels))

	// will call rrdp and sync
	if len(talModels) == 0 {
		return nil
	}

	// get last rrdp logs
	syncRrdpLogs, err := rrdp.GetLastSyncRrdpLogsDb()
	if err != nil {
		belogs.Error("callSync(): rrdp: GetLastSyncRrdpLogsDb fail:", err)
		return
	}

	//start spQueue
	spQueue := NewSyncParseQueue()
	spQueue.LastSyncRrdpLogs = syncRrdpLogs
	spQueue.LabRpkiSyncLogId = syncLogId
	belogs.Debug("callSync(): before startRrdpServer spQueue:", jsonutil.MarshalJson(*spQueue))

	var syncServerWg sync.WaitGroup
	syncServerWg.Add(1)
	go startSyncServer(spQueue, syncState, &syncServerWg)
	belogs.Debug("callSync(): after startSyncServer spQueue:", jsonutil.MarshalJson(*spQueue))

	// start to rrdp by sync url in tal, to get root cer
	// first: remove all root cer, so can will rrdp download and will trigger parse all cer files.
	// otherwise, will have to load all root file manually
	os.RemoveAll(conf.VariableString("rrdp::destPath") + "/root/")
	os.MkdirAll(conf.VariableString("rrdp::destPath")+"/root/", os.ModePerm)
	for _, talModel := range talModels {
		for _, talSyncUrl := range talModel.TalSyncUrls {
			url := ""
			if talSyncUrl.SupportRrdp && len(talSyncUrl.RrdpUrl) > 0 {
				url = talSyncUrl.RrdpUrl
			} else {
				if talSyncUrl.SupportRsync && len(talSyncUrl.RsyncUrl) > 0 {
					url = talSyncUrl.RsyncUrl
				}
			}
			if len(url) > 0 {
				atomic.AddInt64(&spQueue.SyncingAndParsingCount, int64(1))
				belogs.Info("callSync(): will add url:", url, "   current SyncingAndParsingCount:", atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
				go spQueue.AddSyncUrl(url, conf.VariableString("rrdp::destPath")+"/")
			}
		}
	}
	syncServerWg.Wait()
	belogs.Info("callSync():end success: syncState:", jsonutil.MarshalJson(syncState), " time(s):", time.Since(start))
	return nil
}

// call from parseRrdpCerFiles and parseRsyncCerFiles
func parseCerAndGetSubRepoUrlAndAddToSpQueue(spQueue *SyncParseQueue, cerFiles []string) {
	// foreach every cerfiles to parseCerAndGetSubRepoUrl
	belogs.Debug("parseCerAndGetSubRepoUrlAndAddToSpQueue():cerFiles:", cerFiles)

	subRepoUrls := make([]string, 0, len(cerFiles))
	for _, cerFile := range cerFiles {
		// just trigger sync ,no need save to db, ignore err
		subRepoUrl, _ := parseCerAndGetSubRepoUrl(spQueue, cerFile)
		if len(subRepoUrl) > 0 {
			subRepoUrls = append(subRepoUrls, subRepoUrl)
		} else {
			belogs.Error("parseCerAndGetSubRepoUrlAndAddToSpQueue(): this file has no subRepoUrl:", cerFile)
		}
	}
	belogs.Debug("parseCerAndGetSubRepoUrlAndAddToSpQueue():cerFiles:", cerFiles, "  subRepoUrls:", subRepoUrls)

	var syncingAndParsingCountSub int64
	if len(subRepoUrls) == 0 {
		syncingAndParsingCountSub = -1
	} else {
		syncingAndParsingCountSub = int64(len(subRepoUrls)) - 1
	}
	belogs.Info("parseCerAndGetSubRepoUrlAndAddToSpQueue():len(cerFiles):", len(cerFiles),
		"    len(subRepoUrls):", len(subRepoUrls),
		"    syncingAndParsingCountSub:", syncingAndParsingCountSub)

	// the father rsyncingparsingcount -1 ,and the children rsyncingparsingcount + len()
	belogs.Debug("parseCerAndGetSubRepoUrlAndAddToSpQueue():will add subRepoUrls, before SyncingAndParsingCount:",
		atomic.LoadInt64(&spQueue.SyncingAndParsingCount), "  syncingAndParsingCountSub:", syncingAndParsingCountSub)
	atomic.AddInt64(&spQueue.SyncingAndParsingCount, syncingAndParsingCountSub)
	belogs.Debug("parseCerAndGetSubRepoUrlAndAddToSpQueue():will add subRepoUrls, after SyncingAndParsingCount:",
		atomic.LoadInt64(&spQueue.SyncingAndParsingCount), "  syncingAndParsingCountSub:", syncingAndParsingCountSub)

	// call add notifies to rsyncqueue
	if len(subRepoUrls) > 0 {
		addSubRepoUrlsToSpQueue(spQueue, subRepoUrls)
	}
}

// call /parsevalidate/parse to parse cert, and save result
func parseCerAndGetSubRepoUrl(spQueue *SyncParseQueue, cerFile string) (subRepoUrl string, err error) {

	// call parse, not need to save body to db
	start := time.Now()
	belogs.Debug("parseCerAndGetSubRepoUrl():/parsevalidate/parsefilesimple cerFile:", cerFile)
	parseCerSimple := model.ParseCerSimple{}
	err = httpclient.PostFileAndUnmarshalResponseModel("http://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpPort")+
		"/parsevalidate/parsefilesimple", cerFile, "file", false, &parseCerSimple)
	if err != nil {
		spQueue.SyncResult.FailParseValidateCerts.Store(cerFile, err.Error())
		belogs.Error("parseCerAndGetSubRepoUrl(): PostFileAndUnmarshalResponseModel fail:", cerFile, "   err:", err)
		return "", err
	}

	// get the sub repo url in cer, and send it to rpqueue
	subRepoUrl = strings.TrimSpace(parseCerSimple.RpkiNotify)
	if len(subRepoUrl) == 0 {
		subRepoUrl = strings.TrimSpace(parseCerSimple.CaRepository)
	}
	if len(subRepoUrl) == 0 {
		belogs.Error("parseCerAndGetSubRepoUrl(): all rsyncUrl or rrdpUrl is empty:", cerFile, jsonutil.MarshalJson(parseCerSimple))
		return "", errors.New("parseCerAndGetSubRepoUrl(): all rsyncUrl or rrdpUrl is empty:" + cerFile + ",  " + jsonutil.MarshalJson(parseCerSimple))
	}
	belogs.Info("parseCerAndGetSubRepoUrl(): cerFile:", cerFile, "  RpkiNotify:", parseCerSimple.RpkiNotify, "  CaRepository:", parseCerSimple.CaRepository,
		"  subRepoUrl:", subRepoUrl, "  time(s):", time.Since(start))
	belogs.Debug("parseCerAndGetSubRepoUrl(): cerFile:", cerFile, " parseCerSimple:", jsonutil.MarshalJson(parseCerSimple))
	return subRepoUrl, nil

}

func addSubRepoUrlsToSpQueue(spQueue *SyncParseQueue, subRepoUrls []string) {
	rsyncDestPath := conf.VariableString("rsync::destPath") + "/"
	rrdpDestPath := conf.VariableString("rrdp::destPath") + "/"

	belogs.Debug("addSubRepoUrlsToSpQueue(): spQueue.SyncingAndParsingCount+len(subRepoUrls):",
		spQueue.SyncingAndParsingCount, " + ", len(subRepoUrls),
		"		rsyncDestPath:", rsyncDestPath, "   rrdpDestPath:", rrdpDestPath)
	for _, subRepoUrl := range subRepoUrls {
		belogs.Debug("addSubRepoUrlsToSpQueue():will PreCheckSyncUrl, spQueue.SyncingAndParsingCount: ",
			atomic.LoadInt64(&spQueue.SyncingAndParsingCount),
			"   subRepoUrl:", subRepoUrl)
		if !spQueue.PreCheckSyncUrl(subRepoUrl) {
			belogs.Debug("addSubRepoUrlsToSpQueue():PreCheckSyncUrl have exist, before SyncingAndParsingCount-1:", subRepoUrl, atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
			atomic.AddInt64(&spQueue.SyncingAndParsingCount, -1)
			belogs.Debug("addSubRepoUrlsToSpQueue():PreCheckSyncUrl have exist, after SyncingAndParsingCount-1:", subRepoUrl, atomic.LoadInt64(&spQueue.SyncingAndParsingCount))
			continue
		}
		var destPath string
		if strings.Contains(subRepoUrl, "https://") {
			destPath = rrdpDestPath
		} else if strings.Contains(subRepoUrl, "rsync://") {
			destPath = rsyncDestPath
		} else {
			continue
		}
		belogs.Info("addSubRepoUrlsToSpQueue():will AddSyncUrl subRepoUrl: ", subRepoUrl, "  destPath:", destPath)
		go spQueue.AddSyncUrl(subRepoUrl, destPath)
	}
}
