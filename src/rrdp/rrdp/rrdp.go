package rrdp

import (
	"os"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	httpclient "github.com/cpusoft/goutil/httpclient"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"
	rrdputil "github.com/cpusoft/goutil/rrdputil"

	"rrdp/db"
)

// start to rrdp
func Start() {
	belogs.Info("Start():rrdp")

	// save starttime to lab_rpki_sync_log
	syncLogId, err := db.InsertRsyncLogRrdpStateStart("rrdping", "rrdp")
	if err != nil {
		belogs.Error("Start():InsertRsyncLogRsyncStateStart fail:", err)
		return
	}
	belogs.Debug("Start():labRpkiSyncLogId:", syncLogId)

	// get all rpki notify url in cer/conf
	notifyUrls := conf.Strings("rrdp::notifyurls")
	for _, notifyUrl := range notifyUrls {
		rrdpByUrl(notifyUrl, syncLogId)
	}
	err = db.UpdateRsyncLogRrdpStateEnd(syncLogId, "rrdped")
	if err != nil {
		belogs.Error("Start():UpdateRsyncLogRrdpStateEnd fail, syncLogId:",
			syncLogId, err)
		return
	}
	belogs.Info("Start():end rrdp")

	// call parse validate
	go func() {
		httpclient.Post("http", conf.String("rpstir2::parsevalidateserver"), conf.Int("rpstir2::httpport"),
			"/parsevalidate/start", "")
	}()
}

func rrdpByUrl(notifyUrl string, syncLogId uint64) (err error) {

	//defer RsyncByUrlDefer(rsyncUrl, rsyncUrlCh)
	belogs.Debug("rrdpByUrl():start, notifyUrl, syncLogId:", notifyUrl, syncLogId)

	// get notify xml
	notificationModel, err := getRrdpNotification(notifyUrl)
	if err != nil {
		belogs.Error("rrdpByUrl(): GetRrdpNotification fail, notifyUrl, syncLogId: ",
			notifyUrl, syncLogId, err)
		return err
	}

	// get rsync_rrdp_log ,
	has, syncRrdpLog, err := db.GetLastSyncRrdpLog(notifyUrl,
		notificationModel.SessionId,
		notificationModel.MinSerail, notificationModel.MaxSerail)
	if err != nil {
		belogs.Error("rrdpByUrl(): GetLastSyncRrdpLog fail, notifyUrl, syncLogId: ",
			notifyUrl, syncLogId, err)
		return err
	}
	belogs.Debug("rrdpByUrl(): notifyUrl, syncRrdpLog:", notifyUrl,
		jsonutil.MarshalJson(syncRrdpLog))

	// need to get snapshot
	if !has {
		err = processRrdpSnapshot(&notificationModel, syncLogId)

	} else {
		// get delta
		err = processRrdpDelta(&notificationModel, syncLogId, syncRrdpLog.CurSerial)
	}

	if err != nil {
		belogs.Error("rrdpByUrl(): GetLastSyncRrdpLog fail, notifyUrl, syncLogIdsyncLogId: ",
			notifyUrl, syncLogId, err)
		return err
	}

	// insert new sync_rrdp_log
	err = db.InsertSyncRrdpLog(has, syncLogId, notifyUrl, &syncRrdpLog, &notificationModel)
	if err != nil {
		belogs.Error("UpdateRrdpSnapshot():InsertSyncRrdpLog fail: ", err)
		return err
	}
	// get notification xml
	belogs.Debug("rrdpByUrl(): end ok, notifyUrl, syncLogId:", notifyUrl, syncLogId)
	return nil

}

func processRrdpSnapshot(notificationModel *rrdputil.NotificationModel,
	syncRrdpLog uint64) (err error) {

	// first to get snapshot files, because this may fail easily
	snapshotModel, err := getRrdpSnapshot(notificationModel)
	if err != nil {
		belogs.Error("ProcessRrdpSnapshot(): getRrdpSnapshot fail, Snapshot url: ",
			notificationModel.Snapshot.Uri, err)
		return err
	}
	belogs.Debug("processRrdpSnapshot():notificationModel.Snapshot.Uri, snapshotModel:",
		notificationModel.Snapshot.Uri,
		snapshotModel.Serial, len(snapshotModel.SnapshotPublishs))

	// rm disk files
	repoHostPath, err := osutil.GetHostPathFromUrl(conf.VariableString("rrdp::destpath"), notificationModel.Snapshot.Uri)
	if err != nil {
		belogs.Error("ProcessRrdpSnapshot(): GetHostPathFromUrl fail, Snapshot url: ",
			notificationModel.Snapshot.Uri, err)
		return err
	}
	belogs.Debug("processRrdpSnapshot():repoHostPath:", repoHostPath)

	err = os.RemoveAll(repoHostPath)
	if err != nil {
		belogs.Error("ProcessRrdpSnapshot(): RemoveAll, repoHostPath: ", repoHostPath, err)
	}
	err = os.MkdirAll(repoHostPath, os.ModePerm)
	if err != nil {
		belogs.Error("ProcessRrdpSnapshot(): MkdirAll, repoHostPath: ", repoHostPath, err)
	}
	// download snapshot files
	repoPath := conf.VariableString("rrdp::destpath") + osutil.GetPathSeparator()
	err = rrdputil.SaveRrdpSnapshotToFiles(&snapshotModel, repoPath)
	if err != nil {
		belogs.Error("ProcessRrdpSnapshot(): SaveRrdpSnapshotToFiles fail, Snapshot url,  repoPath: ",
			notificationModel.Snapshot.Uri, repoPath, err)
		return err
	}

	// del old cer/crl/mft/roa and update to rsynclog
	err = db.UpdateRrdpSnapshot(&snapshotModel, syncRrdpLog, repoHostPath)
	if err != nil {
		belogs.Error("ProcessRrdpSnapshot(): SaveRrdpSnapshotToFiles fail, Snapshot url,  repoPath: ",
			notificationModel.Snapshot.Uri, repoPath, err)
		return err
	}
	return nil

}

// lastSerial is last syncRrdpLog's curSerial
func processRrdpDelta(notificationModel *rrdputil.NotificationModel,
	syncRrdpLog, lastSerial uint64) (err error) {

	deltaModels, err := getRrdpDelta(notificationModel, lastSerial)
	if err != nil {
		belogs.Error("processRrdpDelta(): getRrdpDelta fail,  len(notificationModel.MapSerialDeltas) :",
			len(notificationModel.MapSerialDeltas), err)
		return err
	}
	belogs.Debug("processRrdpDelta():getRrdpDelta len(deltaModels):", len(deltaModels))
	if len(deltaModels) <= 0 {
		return nil
	}

	// download snapshot files
	repoPath := conf.VariableString("rrdp::destpath") + osutil.GetPathSeparator()
	for i, _ := range deltaModels {
		// save publish files and remove withdraw files
		err = rrdputil.SaveRrdpDeltaToFiles(&deltaModels[i], repoPath)
		if err != nil {
			belogs.Error("processRrdpDelta(): SaveRrdpDeltaToFiles fail, deltaModels[i].Serial,  repoPath: ",
				deltaModels[i].Serial, repoPath, err)
			return err
		}
	}
	for i, _ := range deltaModels {
		// del old cer/crl/mft/roa and update to rsynclog
		// get dest path : /root/rpki/data/reporrdp/
		err = db.UpdateRrdpDelta(&deltaModels[i], syncRrdpLog)
		if err != nil {
			belogs.Error("ProcessRrdpSnapshot(): SaveRrdpSnapshotToFiles fail, Snapshot url,  repoPath: ",
				notificationModel.Snapshot.Uri, repoPath, err)
			return err
		}
	}
	return nil
}
