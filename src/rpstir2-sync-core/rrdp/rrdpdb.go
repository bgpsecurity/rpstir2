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
func updateRrdpSnapshotDb(syncLogId uint64, notificationModel *rrdputil.NotificationModel,
	snapshotModel *rrdputil.SnapshotModel, snapshotDeltaResult *SnapshotDeltaResult,
	syncLogFilesCh chan []model.SyncLogFile) (err error) {

	belogs.Debug("updateRrdpSnapshotDb():syncLogId:", syncLogId,
		"    snapshotDeltaResult:", jsonutil.MarshalJson(snapshotDeltaResult))

	// del cer/crl/mft/roa
	err = sync.DelByFilePathDb(snapshotDeltaResult.RepoHostPath)
	if err != nil {
		belogs.Error("updateRrdpSnapshotDb():delLastRrdpSnapshot fail, repoHostPath:", snapshotDeltaResult.RepoHostPath, err)
		return err
	}

	// insert synclog
	rrdpTime := time.Now()
	syncLogFiles := make([]model.SyncLogFile, 0, 2*len(snapshotDeltaResult.RrdpFiles))
	for _, rrdpFile := range snapshotDeltaResult.RrdpFiles {
		syncLogFile, err := ConvertToSyncLogFile(
			syncLogId, rrdpTime, &rrdpFile)
		if err != nil {
			belogs.Error("updateRrdpSnapshotDb():ConvertToSyncLogFile fail, rrdpFile:", jsonutil.MarshalJson(rrdpFile), err)
			return err
		}
		syncLogFiles = append(syncLogFiles, syncLogFile)
	}

	belogs.Debug("updateRrdpSnapshotDb():rrdp, len(syncLogFiles):", len(syncLogFiles),
		"  len(rrdpFile):", len(snapshotDeltaResult.RrdpFiles))
	if syncLogFilesCh != nil {
		belogs.Debug("updateRrdpSnapshotDb():rrdp len(syncLogFiles) -> syncLogFilesCh:", len(syncLogFiles))
		syncLogFilesCh <- syncLogFiles
	} else {
		belogs.Debug("updateRrdpSnapshotDb():rrdp InsertSyncLogFilesDb():", len(syncLogFiles))
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
	err = insertSyncRrdpLogDb(session, syncLogId, snapshotDeltaResult)
	if err != nil {
		belogs.Error("updateRrdpSnapshotDb():insertSyncRrdpLogDb fail, syncLogId, notifyUrl:",
			syncLogId, snapshotDeltaResult.NotifyUrl, err)
		return xormdb.RollbackAndLogError(session, "updateRrdpSnapshotDb(): CommitSession fail:", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "updateRrdpSnapshotDb(): CommitSession fail:", err)
	}
	return nil
}

//
func updateRrdpDeltaDb(syncLogId uint64, deltaModels []rrdputil.DeltaModel,
	snapshotDeltaResult *SnapshotDeltaResult, syncLogFilesCh chan []model.SyncLogFile) (err error) {
	belogs.Debug("updateRrdpDeltaDb():syncLogId :", syncLogId, "    snapshotDeltaResult:", jsonutil.MarshalJson(snapshotDeltaResult))

	// insert synclog
	rrdpTime := time.Now()
	syncLogFiles := make([]model.SyncLogFile, 0, 2*len(snapshotDeltaResult.RrdpFiles))
	for _, rrdpFile := range snapshotDeltaResult.RrdpFiles {
		syncLogFile, err := ConvertToSyncLogFile(
			syncLogId, rrdpTime, &rrdpFile)
		if err != nil {
			belogs.Error("updateRrdpDeltaDb():ConvertToSyncLogFile fail, rrdpFile:", jsonutil.MarshalJson(rrdpFile), err)
			return err
		}
		syncLogFiles = append(syncLogFiles, syncLogFile)
	}
	belogs.Debug("updateRrdpDeltaDb():rrdp, len(syncLogFiles):", len(syncLogFiles),
		"  len(rrdpFile):", len(snapshotDeltaResult.RrdpFiles))
	if syncLogFilesCh != nil {
		belogs.Debug("updateRrdpDeltaDb():rrdp len(syncLogFiles) -> syncLogFilesCh:", len(syncLogFiles))
		syncLogFilesCh <- syncLogFiles
	} else {
		belogs.Debug("updateRrdpDeltaDb():rrdp InsertSyncLogFilesDb():", len(syncLogFiles))
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
		err = insertSyncRrdpLogDb(session, syncLogId, snapshotDeltaResult)
		if err != nil {
			belogs.Error("updateRrdpDeltaDb():insertSyncRrdpLogDb fail, syncLogId, notifyUrl:",
				syncLogId, snapshotDeltaResult.NotifyUrl, err)
			return xormdb.RollbackAndLogError(session, "updateRrdpSnapshotDb(): CommitSession fail:", err)
		}
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "updateRrdpDeltaDb(): CommitSession fail:", err)
	}
	return nil
}
