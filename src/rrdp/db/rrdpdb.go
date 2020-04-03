package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	conf "github.com/cpusoft/goutil/conf"
	osutil "github.com/cpusoft/goutil/osutil"
	rrdputil "github.com/cpusoft/goutil/rrdputil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"
)

// repoHostPath, is nic dest path, eg: /root/rpki/data/reporrdp/rpki.apnic.cn/
func UpdateRrdpSnapshot(snapshotModel *rrdputil.SnapshotModel,
	labRpkiSyncLogId uint64, repoHostPath string) (err error) {

	belogs.Debug("UpdateRrdpSnapshot():labRpkiSyncLogId,repoHostPath :",
		labRpkiSyncLogId, repoHostPath)

	// delete in cer/crl/mft/roa table
	session, err := xormdb.NewSession()
	defer session.Close()

	// del cer/crl/mft/roa
	err = delLastRrdpSnapshot(session, repoHostPath)
	if err != nil {
		belogs.Error("UpdateRrdpSnapshot():delLastRrdpSnapshot fail, repoHostPath:", repoHostPath, err)
		return xormdb.RollbackAndLogError(session, "delLastRrdpSnapshot fail, repoHostPath:"+repoHostPath, err)
	}

	// insert synclog
	rrdpTime := time.Now()
	for i, _ := range snapshotModel.SnapshotPublishs {

		pathFileName, err := osutil.GetPathFileNameFromUrl(conf.VariableString("rrdp::destpath"), snapshotModel.SnapshotPublishs[i].Uri)
		if err != nil {
			belogs.Error("UpdateRrdpSnapshot(): GetPathFileNameFromUrl fail:", snapshotModel.SnapshotPublishs[i].Uri)
			return err
		}
		belogs.Debug("UpdateRrdpSnapshot():before InsertRsyncLogFile: labRpkiSyncLogId,rrdpTime,pathFileName:",
			labRpkiSyncLogId, rrdpTime, pathFileName)

		err = InsertRsyncLogFile(session,
			labRpkiSyncLogId,
			snapshotModel.SnapshotPublishs[i].Uri, "add", pathFileName,
			rrdpTime)
		if err != nil {
			belogs.Error("UpdateRrdpSnapshot():InsertRsyncLogFile fail, labRpkiSyncLogId,rrdpTime,pathFileName:",
				labRpkiSyncLogId, rrdpTime, pathFileName, err)
			return xormdb.RollbackAndLogError(session, "ProcessRrdpSnapshot(): CommitSession fail:", err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "ProcessRrdpSnapshot(): CommitSession fail:", err)
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
func UpdateRrdpDelta(deltaModel *rrdputil.DeltaModel,
	labRpkiSyncLogId uint64) (err error) {

	belogs.Debug("UpdateRrdpDelta():labRpkiSyncLogId :",
		labRpkiSyncLogId)

	// delete in cer/crl/mft/roa table
	session, err := xormdb.NewSession()
	defer session.Close()

	// insert synclog
	rrdpTime := time.Now()
	// first to withdraw
	for i, _ := range deltaModel.DeltaWithdraws {

		pathFileName, err := osutil.GetPathFileNameFromUrl(conf.VariableString("rrdp::destpath"), deltaModel.DeltaWithdraws[i].Uri)
		if err != nil {
			belogs.Error("UpdateRrdpSnapshot():DeltaWithdraws GetPathFileNameFromUrl fail:", deltaModel.DeltaWithdraws[i].Uri)
			return err
		}
		belogs.Debug("UpdateRrdpDelta():DeltaWithdrawsbefore InsertRsyncLogFile: labRpkiSyncLogId,rrdpTime,pathFileName:",
			labRpkiSyncLogId, rrdpTime, pathFileName)

		err = InsertRsyncLogFile(session,
			labRpkiSyncLogId,
			deltaModel.DeltaWithdraws[i].Uri, "del", pathFileName,
			rrdpTime)
		if err != nil {
			belogs.Error("UpdateRrdpDelta():DeltaWithdrawsInsertRsyncLogFile fail, labRpkiSyncLogId,rrdpTime,pathFileName:",
				labRpkiSyncLogId, rrdpTime, pathFileName, err)
			return xormdb.RollbackAndLogError(session, "UpdateRrdpDelta():DeltaWithdraws CommitSession fail:", err)
		}
	}

	for i, _ := range deltaModel.DeltaPublishs {

		pathFileName, err := osutil.GetPathFileNameFromUrl(conf.VariableString("rrdp::destpath"), deltaModel.DeltaPublishs[i].Uri)
		if err != nil {
			belogs.Error("UpdateRrdpSnapshot():DeltaPublishs GetPathFileNameFromUrl fail:", deltaModel.DeltaPublishs[i].Uri)
			return err
		}
		belogs.Debug("UpdateRrdpDelta():DeltaPublishsbefore InsertRsyncLogFile: labRpkiSyncLogId,rrdpTime,pathFileName:",
			labRpkiSyncLogId, rrdpTime, pathFileName)

		err = InsertRsyncLogFile(session,
			labRpkiSyncLogId,
			deltaModel.DeltaPublishs[i].Uri, "add", pathFileName,
			rrdpTime)
		if err != nil {
			belogs.Error("UpdateRrdpDelta():DeltaPublishsInsertRsyncLogFile fail, labRpkiSyncLogId,rrdpTime,filePathName:",
				labRpkiSyncLogId, rrdpTime, pathFileName, err)
			return xormdb.RollbackAndLogError(session, "UpdateRrdpDelta():DeltaPublishs CommitSession fail:", err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "UpdateRrdpDelta(): CommitSession fail:", err)
	}
	return nil
}
