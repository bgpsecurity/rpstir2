package sync

import (
	"errors"
	"time"

	model "rpstir2-model"
	"rpstir2-sync-core/sync"
	rpsync "rpstir2-sync-core/sync"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/urlutil"
)

var rrdpResultCh chan model.SyncResult
var rsyncResultCh chan model.SyncResult

func init() {
	rrdpResultCh = make(chan model.SyncResult)
	rsyncResultCh = make(chan model.SyncResult)
	belogs.Debug("init(): chan rrdpResultCh:", rrdpResultCh, "   chan rsyncResultCh:", rsyncResultCh)
}

func syncStart(syncStyle model.SyncStyle) (nextStep string, err error) {
	start := time.Now()

	belogs.Info("syncStart():syncStyle:", syncStyle)

	syncLogSyncState := model.SyncLogSyncState{StartTime: time.Now(), SyncStyle: syncStyle.SyncStyle}

	// syncStyle: sync/rsync/rrdp,state: syncing;
	syncLogId, err := sync.InsertSyncLogStartDb(syncStyle.SyncStyle, "syncing")
	if err != nil {
		belogs.Error("syncStart():InsertSyncLogStartDb fail:", err)
		return "", err
	}
	belogs.Info("syncStart():syncLogId:", syncLogId, "  syncLogSyncState:", jsonutil.MarshalJson(syncLogSyncState))

	// call tals , get all tals
	talModels, err := getTals()
	if err != nil {
		belogs.Error("syncStart(): getTals failed, err:", err)
		return "", err
	}
	belogs.Debug("syncStart(): len(talModels):", len(talModels))

	// classify rsync and rrdp
	syncLogSyncState.RrdpUrls, syncLogSyncState.RsyncUrls, err = getUrlsBySyncStyle(syncStyle, talModels)
	if err != nil {
		belogs.Error("syncStart(): getUrlsBySyncType fail")
		return "", err
	}
	belogs.Debug("syncStart(): rrdpUrls:", syncLogSyncState.RrdpUrls, "   rsyncUrls:", syncLogSyncState.RsyncUrls)

	// Check whether this time sync mode is different from the last sync mode.
	// it means actual directory is different from this sync direcotry.
	// for example, if actual directory is sync , but this time sync mode is rrdp
	// then there must be full sync
	needFullSync, err := checkNeedFullSync(syncLogSyncState.RrdpUrls, syncLogSyncState.RsyncUrls)
	if needFullSync || err != nil {
		belogs.Debug("syncStart(): checkNeedFullSync fail, rrdpUrls: ", syncLogSyncState.RrdpUrls, "   rsyncUrls:", syncLogSyncState.RsyncUrls, err)
		belogs.Info("syncStart(): because this time sync mode is different from the last sync mode, so  a full sync has to be triggered")
		return "fullsync", nil

	}

	// call rrdp and rsync and wait for result
	err = callRrdpAndRsync(syncLogId, &syncLogSyncState)
	if err != nil {
		belogs.Error("syncStart():callRrdpAndRsync fail:", err)
		return "", err
	}

	// update lab_rpki_sync_log
	err = rpsync.UpdateSyncLogEndDb(syncLogId, "synced", jsonutil.MarshalJson(syncLogSyncState))
	if err != nil {
		belogs.Error("syncStart():UpdateSyncLogEndDb fail:", err)
		return "", err
	}
	belogs.Info("syncStart(): end sync, will parsevalidate,  time(s):", time.Now().Sub(start).Seconds())

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
		jsonutil.MarshalJson(talModelsResponse), "  time(s):", time.Now().Sub(start).Seconds())

	if len(talModelsResponse.TalModels) == 0 {
		belogs.Error("getTals(): there is no tal file")
		return nil, errors.New("there is no tal file")
	}
	return talModelsResponse.TalModels, nil
}

func getUrlsBySyncStyle(syncStyle model.SyncStyle, talModels []model.TalModel) (rrdpUrls, rsyncUrls []string, err error) {
	belogs.Debug("getUrlsBySyncStyle(): syncStyle:", syncStyle, "      talModels:", jsonutil.MarshalJson(talModels))
	for i := range talModels {

		for j := range talModels[i].TalSyncUrls {
			if syncStyle.SyncStyle == "sync" {
				if talModels[i].TalSyncUrls[j].SupportRrdp {
					rrdpUrls = append(rrdpUrls, talModels[i].TalSyncUrls[j].RrdpUrl)
				} else if talModels[i].TalSyncUrls[j].SupportRsync {
					rsyncUrls = append(rsyncUrls, talModels[i].TalSyncUrls[j].RsyncUrl)
				}
			} else if syncStyle.SyncStyle == "rrdp" {
				if talModels[i].TalSyncUrls[j].SupportRrdp {
					rrdpUrls = append(rrdpUrls, talModels[i].TalSyncUrls[j].RrdpUrl)
				}
			} else if syncStyle.SyncStyle == "rsync" {
				if talModels[i].TalSyncUrls[j].SupportRsync {
					rsyncUrls = append(rsyncUrls, talModels[i].TalSyncUrls[j].RsyncUrl)
				}
			}
		}
	}
	belogs.Debug("getUrlsBySyncStyle(): syncStyle:", syncStyle,
		"      rrdpUrls:", rrdpUrls, "  rsyncUrls:", rsyncUrls)

	if len(rrdpUrls) == 0 && len(rsyncUrls) == 0 {
		belogs.Error("getUrlsBySyncType(): there is neighor rrdp urls nor rsync urls")
		return nil, nil, errors.New("there is neighor rrdp urls nor rsync urls")
	}

	return
}

func checkNeedFullSync(thisRrdpUrls, thisRsyncUrls []string) (needFullSync bool, err error) {
	needFullSync = false
	rrdpDestPath := conf.VariableString("rrdp::destPath") + osutil.GetPathSeparator()
	rsyncDestPath := conf.VariableString("rsync::destPath") + osutil.GetPathSeparator()
	belogs.Debug("checkNeedFullSync(): rrdpDestPath,  rsyncDestPath:", rrdpDestPath, rsyncDestPath,
		"  thisRrdpUrls:", thisRrdpUrls, "     thisRsyncUrls:", thisRsyncUrls)

	// if rrdp url exists in sync, or sync url exists in rrdp, it will needFullSync
	for _, thisRrdpUrl := range thisRrdpUrls {
		testRrdpUrlInRsyncDestPath, err := urlutil.JoinPrefixPathAndUrlHost(rsyncDestPath, thisRrdpUrl)
		belogs.Debug("checkNeedFullSync(): test rrdp url in sync:", testRrdpUrlInRsyncDestPath)
		if err != nil {
			belogs.Error("checkNeedFullSync():test rrdp url exists in rsync, JoinPrefixPathAndUrlHost err,  rsyncDestPath, thisRrdpUrl:", rsyncDestPath, thisRrdpUrl)
			return true, err
		}
		exists, err := osutil.IsExists(testRrdpUrlInRsyncDestPath)
		if err != nil {
			belogs.Info("checkNeedFullSync(): test rrdp url exists in rsync, IsExists err, testRrdpUrlInRsyncDestPath:", testRrdpUrlInRsyncDestPath, err)
			return true, err
		}
		if exists {
			belogs.Debug("checkNeedFullSync(): test rrdp url exists in rsync, need full sync:", testRrdpUrlInRsyncDestPath)
			return true, nil
		}
	}
	for _, thisRsyncUrl := range thisRsyncUrls {
		testRsyncUrlInRrdpDestPath, err := urlutil.JoinPrefixPathAndUrlHost(rrdpDestPath, thisRsyncUrl)
		belogs.Debug("checkNeedFullSync(): test rsync url in rrdp:", testRsyncUrlInRrdpDestPath)
		if err != nil {
			belogs.Error("checkNeedFullSync(): test rsync url exists in rrdp, JoinPrefixPathAndUrlHost err,  rrdpDestPath, thisRsyncUrl:", rrdpDestPath, thisRsyncUrl)
			return true, err
		}
		exists, err := osutil.IsExists(testRsyncUrlInRrdpDestPath)
		if err != nil {
			belogs.Error("checkNeedFullSync(): test rsync exists in rrdp, IsExists err, testRsyncUrlInRrdpDestPath:", testRsyncUrlInRrdpDestPath, err)
			return true, err
		}
		if exists {
			belogs.Info("checkNeedFullSync(): test rsync url exits in rrdp ,need full sync:", testRsyncUrlInRrdpDestPath)
			return true, nil
		}
	}
	belogs.Debug("checkNeedFullSync(): not need full sync")
	return false, nil
}

func callRrdpAndRsync(syncLogId uint64, syncLogSyncState *model.SyncLogSyncState) (err error) {

	syncUrls := model.SyncUrls{
		SyncLogId: syncLogId,
		RrdpUrls:  syncLogSyncState.RrdpUrls,
		RsyncUrls: syncLogSyncState.RsyncUrls}
	syncUrlsJson := jsonutil.MarshalJson(syncUrls)
	belogs.Debug("callRrdpAndRsync(): syncUrlsJson:", syncUrlsJson)

	// if there is no rrdp ,then rrdpEnd=true. same to rsyncEnd
	rrdpEnd := false
	rsyncEnd := false
	// will call rrdp and sync
	if len(syncUrls.RrdpUrls) > 0 {
		go func() {
			httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
				"/entiresync/rrdpstart", syncUrlsJson, false)
		}()
	} else {
		rrdpEnd = true
	}

	if len(syncUrls.RsyncUrls) > 0 {
		go func() {
			httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
				"/entiresync/rsyncstart", syncUrlsJson, false)
		}()
	} else {
		rsyncEnd = true
	}

	// both rrdpEnd==true and rsyncEnd==true, will end select
	belogs.Debug("callRrdpAndRsync(): rrdpEnd, rsyncEnd:", rrdpEnd, rsyncEnd,
		" chan rrdpResultCh:", rrdpResultCh, "   chan rsyncResultCh:", rsyncResultCh)
	for {
		select {
		case syncLogSyncState.RrdpResult = <-rrdpResultCh:
			belogs.Debug("callRrdpAndRsync(): rrdpResult:", jsonutil.MarshalJson(syncLogSyncState.RrdpResult))
			rrdpEnd = true
		case syncLogSyncState.RsyncResult = <-rsyncResultCh:
			belogs.Debug("callRrdpAndRsync(): rsyncResult:", jsonutil.MarshalJson(syncLogSyncState.RsyncResult))
			rsyncEnd = true
		}
		if rrdpEnd && rsyncEnd {
			belogs.Debug("callRrdpAndRsync(): for select  end")
			break
		}
	}
	syncLogSyncState.EndTime = time.Now()
	belogs.Debug("callRrdpAndRsync(): end")
	return
}
func rrdpResult(r *model.SyncResult) {
	belogs.Debug("RrdpResult(): get syncResult:", jsonutil.MarshalJson(*r), "   chan rrdpResultCh:", rrdpResultCh)
	rrdpResultCh <- *r

}
func rsyncResult(r *model.SyncResult) {
	belogs.Debug("RsyncResult(): get syncResult:", jsonutil.MarshalJson(*r), "   chan rsyncResultCh:", rsyncResultCh)
	rsyncResultCh <- *r

}

func LocalStart() {
	start := time.Now()

	// local sync will set as rsync
	belogs.Info("LocalStart():syncStyle:  rsync")
	syncLogSyncState := model.SyncLogSyncState{StartTime: time.Now(), SyncStyle: "rsync"}

	// start , insert lab_rpki_sync_log

	syncLogId, err := sync.InsertSyncLogStartDb("rsync", "syncing")
	if err != nil {
		belogs.Error("LocalStart():InsertSyncLogSyncStateStart fail:", err)
		return
	}
	belogs.Info("LocalStart():syncLogId:", syncLogId, "  syncLogSyncState:", jsonutil.MarshalJson(syncLogSyncState))

	// call local such as rsync and wait for result
	err = callLocalRsync(syncLogId, &syncLogSyncState)
	if err != nil {
		belogs.Error("LocalStart():callLocalRsync fail:", err)
		return
	}
	belogs.Debug("LocalStart(): end callLocalRsync:", jsonutil.MarshalJson(syncLogSyncState))

	// update lab_rpki_sync_log
	err = sync.UpdateSyncLogEndDb(syncLogId, "synced", jsonutil.MarshalJson(syncLogSyncState))
	if err != nil {
		belogs.Error("LocalStart():UpdateSyncLogEndDb fail:", err)
		return
	}

	// leave serviceState
	httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
		"/sys/servicestate", `{"operate":"leave","state":"sync"}`, false)

	belogs.Info("LocalStart(): sync end , will call parsevalidate,  time(s):", time.Now().Sub(start).Seconds())
	// will call parseValidate
	go func() {
		httpclient.Post("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
			"/parsevalidate/start", "", false)
	}()

}

func callLocalRsync(syncLogId uint64, syncLogSyncState *model.SyncLogSyncState) (err error) {

	syncUrls := model.SyncUrls{
		SyncLogId: syncLogId}
	syncUrlsJson := jsonutil.MarshalJson(syncUrls)
	belogs.Debug("callLocalRsync(): syncUrlsJson:", syncUrlsJson)
	rsyncResult := model.SyncResult{}
	httpclient.SetTimeout(30)
	defer httpclient.ResetTimeout()
	err = httpclient.PostAndUnmarshalResponseModel("https://"+conf.String("rpstir2-rp::serverHost")+":"+conf.String("rpstir2-rp::serverHttpsPort")+
		"/entiresync/rsynclocalstart", syncUrlsJson, false, &rsyncResult)
	if err != nil {
		belogs.Error("callLocalRsync(): rsync localstart failed:", syncUrlsJson, "  err:", err)
		return err
	}
	belogs.Debug("callLocalRsync():after /entiresync/rsynclocalstart, syncUrlsJson:", syncUrlsJson, "   rsyncResult:", jsonutil.MarshalJson(rsyncResult))

	syncLogSyncState.RsyncResult = rsyncResult
	syncLogSyncState.EndTime = time.Now()
	belogs.Debug("callLocalRsync(): end")
	return
}
