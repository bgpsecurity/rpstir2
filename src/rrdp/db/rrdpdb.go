package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	hashutil "github.com/cpusoft/goutil/hashutil"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"
	rrdputil "github.com/cpusoft/goutil/rrdputil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"

	rrdpmodel "rrdp/model"
)

// repoHostPath, is nic dest path, eg: /root/rpki/data/reporrdp/rpki.apnic.cn/
func UpdateRrdpSnapshot(syncLogId uint64, notificationModel *rrdputil.NotificationModel, snapshotModel *rrdputil.SnapshotModel,
	snapshotDeltaResult *rrdpmodel.SnapshotDeltaResult) (err error) {

	belogs.Debug("UpdateRrdpSnapshot():syncLogId,  snapshotDeltaResult:", syncLogId, jsonutil.MarshalJson(snapshotDeltaResult))

	// delete in cer/crl/mft/roa table
	session, err := xormdb.NewSession()
	defer session.Close()

	// del cer/crl/mft/roa
	err = delLastRrdpSnapshot(session, snapshotDeltaResult.RepoHostPath)
	if err != nil {
		belogs.Error("UpdateRrdpSnapshot():delLastRrdpSnapshot fail, repoHostPath:", snapshotDeltaResult.RepoHostPath, err)
		return xormdb.RollbackAndLogError(session, "delLastRrdpSnapshot fail, repoHostPath:"+snapshotDeltaResult.RepoHostPath, err)
	}

	// insert synclog
	rrdpTime := time.Now()
	for i := range snapshotDeltaResult.RrdpFiles {
		file := osutil.JoinPathFile(snapshotDeltaResult.RrdpFiles[i].FilePath, snapshotDeltaResult.RrdpFiles[i].FileName)
		fileHash, err := hashutil.Sha256File(file)
		if err != nil {
			belogs.Error("UpdateRrdpSnapshot():Sha256File fail, file:", file, err)
			return xormdb.RollbackAndLogError(session, "get file hash fail, file:"+file, err)
		}
		err = InsertSyncLogFile(session,
			syncLogId,
			"add", snapshotDeltaResult.RrdpFiles[i].FilePath, snapshotDeltaResult.RrdpFiles[i].FileName,
			fileHash,
			rrdpTime)
		if err != nil {
			belogs.Error("UpdateRrdpSnapshot():InsertSyncLogFile fail, syncLogId,rrdpFiles[i].FilePath, rrdpFiles[i].FileName:",
				syncLogId, snapshotDeltaResult.RrdpFiles[i].FilePath, snapshotDeltaResult.RrdpFiles[i].FileName, err)
			return xormdb.RollbackAndLogError(session, "UpdateRrdpSnapshot(): InsertRsyncLogFile fail:", err)
		}
	}
	snapshotDeltaResult.SessionId = notificationModel.SessionId
	snapshotDeltaResult.Serial = notificationModel.Serial
	snapshotDeltaResult.LastSerial = 0
	snapshotDeltaResult.RrdpType = "snapshot"
	snapshotDeltaResult.RrdpTime = rrdpTime
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

// repoHostPath, is nic dest path, eg: /root/rpki/data/reporrdp/rpki.apnic.cn/
func delLastRrdpSnapshot(session *xorm.Session, repoHostPath string) (err error) {

	belogs.Debug("delLastRrdpSnapshot(): repoHostPath:", repoHostPath)

	err = DelCer(session, repoHostPath)
	if err != nil {
		belogs.Error("ProcessRrdpSnapshot(): DelCer fail, repoHostPath: ",
			repoHostPath, err)
		return err
	}

	err = DelCrl(session, repoHostPath)
	if err != nil {
		belogs.Error("ProcessRrdpSnapshot(): DelCrl fail, repoHostPath: ",
			repoHostPath, err)
		return err
	}

	err = DelMft(session, repoHostPath)
	if err != nil {
		belogs.Error("ProcessRrdpSnapshot(): DelMft fail, repoHostPath: ",
			repoHostPath, err)
		return err
	}

	err = DelRoa(session, repoHostPath)
	if err != nil {
		belogs.Error("ProcessRrdpSnapshot(): DelRoa fail, repoHostPath: ",
			repoHostPath, err)
		return err
	}
	return nil
}

//
func UpdateRrdpDelta(syncLogId uint64, deltaModels []rrdputil.DeltaModel, snapshotDeltaResult *rrdpmodel.SnapshotDeltaResult) (err error) {

	belogs.Debug("UpdateRrdpDelta():syncLogId :", syncLogId)

	// delete in cer/crl/mft/roa table
	session, err := xormdb.NewSession()
	defer session.Close()

	// insert synclog
	rrdpTime := time.Now()
	for i := range snapshotDeltaResult.RrdpFiles {
		fileHash := ""
		if snapshotDeltaResult.RrdpFiles[i].SyncType == "add" {
			file := osutil.JoinPathFile(snapshotDeltaResult.RrdpFiles[i].FilePath, snapshotDeltaResult.RrdpFiles[i].FileName)
			fileHash, err = hashutil.Sha256File(file)
			if err != nil {
				belogs.Error("UpdateRrdpDelta():Sha256File fail, file:", file, err)
				return xormdb.RollbackAndLogError(session, "get file hash fail, file:"+file, err)
			}
		}
		err = InsertSyncLogFile(session,
			syncLogId,
			snapshotDeltaResult.RrdpFiles[i].SyncType, snapshotDeltaResult.RrdpFiles[i].FilePath,
			snapshotDeltaResult.RrdpFiles[i].FileName, fileHash,
			rrdpTime)
		if err != nil {
			belogs.Error("UpdateRrdpDelta():InsertSyncLogFile fail, syncLogId,rrdpFiles[i].FilePath, rrdpFiles[i].FileName:",
				syncLogId, snapshotDeltaResult.RrdpFiles[i].FilePath, snapshotDeltaResult.RrdpFiles[i].FileName, err)
			return xormdb.RollbackAndLogError(session, "UpdateRrdpDelta(): InsertRsyncLogFile fail:", err)
		}
	}
	for i := range deltaModels {
		snapshotDeltaResult.SessionId = deltaModels[i].SessionId
		snapshotDeltaResult.Serial = deltaModels[i].Serial
		snapshotDeltaResult.LastSerial = snapshotDeltaResult.LastSerial
		snapshotDeltaResult.RrdpType = "delta"
		snapshotDeltaResult.RrdpTime = rrdpTime
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
