package rrdp

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/rrdputil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
	"rpstir2-sync-core/sync"
)

// repoHostPath, is nic dest path, eg: /root/rpki/data/reporrdp/rpki.apnic.cn/
func UpdateRrdpSnapshot(syncLogId uint64, notificationModel *rrdputil.NotificationModel,
	snapshotModel *rrdputil.SnapshotModel, snapshotDeltaResult *SnapshotDeltaResult,
	syncLogFilesCh chan []model.SyncLogFile) (err error) {

	belogs.Debug("UpdateRrdpSnapshot():syncLogId:", syncLogId,
		"    snapshotDeltaResult:", jsonutil.MarshalJson(snapshotDeltaResult))

	// del cer/crl/mft/roa
	err = sync.DelByFilePathDb(snapshotDeltaResult.RepoHostPath)
	if err != nil {
		belogs.Error("UpdateRrdpSnapshot():delLastRrdpSnapshot fail, repoHostPath:", snapshotDeltaResult.RepoHostPath, err)
		return err
	}

	// insert synclog
	rrdpTime := time.Now()
	syncLogFiles := make([]model.SyncLogFile, 0, 2*len(snapshotDeltaResult.RrdpFiles))
	for _, rrdpFile := range snapshotDeltaResult.RrdpFiles {
		syncLogFile, err := ConvertToSyncLogFile(
			syncLogId, rrdpTime, &rrdpFile)
		if err != nil {
			belogs.Error("UpdateRrdpSnapshot():ConvertToSyncLogFile fail, rrdpFile:", jsonutil.MarshalJson(rrdpFile), err)
			return err
		}
		syncLogFiles = append(syncLogFiles, syncLogFile)
	}

	belogs.Debug("UpdateRrdpSnapshot():rrdp, len(syncLogFiles):", len(syncLogFiles),
		"  len(rrdpFile):", len(snapshotDeltaResult.RrdpFiles))
	if syncLogFilesCh != nil {
		belogs.Debug("UpdateRrdpSnapshot():rrdp len(syncLogFiles) -> syncLogFilesCh:", len(syncLogFiles))
		syncLogFilesCh <- syncLogFiles
	} else {
		belogs.Debug("UpdateRrdpSnapshot():rrdp InsertSyncLogFilesDb():", len(syncLogFiles))
		sync.InsertSyncLogFilesDb(syncLogFiles)
	}
	// delete in cer/crl/mft/roa table
	session, err := xormdb.NewSession()
	defer session.Close()
	snapshotDeltaResult.SessionId = notificationModel.SessionId
	snapshotDeltaResult.Serial = notificationModel.Serial
	snapshotDeltaResult.LastSerial = 0
	snapshotDeltaResult.RrdpType = "snapshot"
	snapshotDeltaResult.RrdpTime = rrdpTime
	snapshotDeltaResult.SnapshotOrDeltaUrl = snapshotModel.SnapshotUrl
	err = InsertSyncRrdpLog(session, syncLogId, snapshotDeltaResult)
	if err != nil {
		belogs.Error("UpdateRrdpSnapshot():InsertSyncRrdpLog fail, syncLogId, notifyUrl:",
			syncLogId, snapshotDeltaResult.NotifyUrl, err)
		return xormdb.RollbackAndLogError(session, "UpdateRrdpSnapshot(): CommitSession fail:", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRrdpSnapshot(): CommitSession fail:", err)
	}
	return nil
}

//
func UpdateRrdpDelta(syncLogId uint64, deltaModels []rrdputil.DeltaModel,
	snapshotDeltaResult *SnapshotDeltaResult, syncLogFilesCh chan []model.SyncLogFile) (err error) {
	belogs.Debug("UpdateRrdpDelta():syncLogId :", syncLogId, "    snapshotDeltaResult:", jsonutil.MarshalJson(snapshotDeltaResult))

	// insert synclog
	rrdpTime := time.Now()
	syncLogFiles := make([]model.SyncLogFile, 0, 2*len(snapshotDeltaResult.RrdpFiles))
	for _, rrdpFile := range snapshotDeltaResult.RrdpFiles {
		syncLogFile, err := ConvertToSyncLogFile(
			syncLogId, rrdpTime, &rrdpFile)
		if err != nil {
			belogs.Error("UpdateRrdpDelta():ConvertToSyncLogFile fail, rrdpFile:", jsonutil.MarshalJson(rrdpFile), err)
			return err
		}
		syncLogFiles = append(syncLogFiles, syncLogFile)
	}
	belogs.Debug("UpdateRrdpDelta():rrdp, len(syncLogFiles):", len(syncLogFiles),
		"  len(rrdpFile):", len(snapshotDeltaResult.RrdpFiles))
	if syncLogFilesCh != nil {
		belogs.Debug("UpdateRrdpDelta():rrdp len(syncLogFiles) -> syncLogFilesCh:", len(syncLogFiles))
		syncLogFilesCh <- syncLogFiles
	} else {
		belogs.Debug("UpdateRrdpDelta():rrdp InsertSyncLogFilesDb():", len(syncLogFiles))
		sync.InsertSyncLogFilesDb(syncLogFiles)
	}

	// insert lab_rpki_sync_rrdp_log table
	session, err := xormdb.NewSession()
	defer session.Close()
	for i := range deltaModels {
		snapshotDeltaResult.SessionId = deltaModels[i].SessionId
		snapshotDeltaResult.Serial = deltaModels[i].Serial
		snapshotDeltaResult.LastSerial = snapshotDeltaResult.LastSerial
		snapshotDeltaResult.RrdpType = "delta"
		snapshotDeltaResult.RrdpTime = rrdpTime
		snapshotDeltaResult.SnapshotOrDeltaUrl = deltaModels[i].DeltaUrl
		err = InsertSyncRrdpLog(session, syncLogId, snapshotDeltaResult)
		if err != nil {
			belogs.Error("UpdateRrdpDelta():InsertSyncRrdpLog fail, syncLogId, notifyUrl:",
				syncLogId, snapshotDeltaResult.NotifyUrl, err)
			return xormdb.RollbackAndLogError(session, "UpdateRrdpSnapshot(): CommitSession fail:", err)
		}
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRrdpDelta(): CommitSession fail:", err)
	}
	return nil
}
