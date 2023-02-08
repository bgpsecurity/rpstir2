package rrdp

import (
	"os"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/rrdputil"
	"github.com/cpusoft/goutil/urlutil"
)

func getRrdpSnapshot(notificationModel *rrdputil.NotificationModel) (snapshotModel rrdputil.SnapshotModel, err error) {

	belogs.Debug("getRrdpSnapshot(): Snapshot.Uri :", notificationModel.Snapshot.Uri)
	snapshotModel, err = rrdputil.GetRrdpSnapshot(notificationModel.Snapshot.Uri)
	if err != nil {
		belogs.Error("getRrdpSnapshot(): GetRrdpSnapshot fail,Snapshot.Uri :",
			notificationModel.Snapshot.Uri, err)
		return snapshotModel, err
	}

	err = rrdputil.CheckRrdpSnapshot(&snapshotModel, notificationModel)
	if err != nil {
		belogs.Error("getRrdpSnapshot(): CheckRrdpSnapshot fail, Snapshot.Uri,snapshotModel :",
			notificationModel.Snapshot.Uri, err)
		belogs.Debug("getRrdpSnapshot(): CheckRrdpSnapshot fail, Snapshot.Uri,snapshotModel :",
			notificationModel.Snapshot.Uri, jsonutil.MarshalJson(snapshotModel), err)
		return snapshotModel, err
	}
	return snapshotModel, nil

}

func processRrdpSnapshot(syncLogId uint64, notificationModel *rrdputil.NotificationModel,
	snapshotDeltaResult *SnapshotDeltaResult, syncLogFilesCh chan []model.SyncLogFile) (err error) {

	belogs.Debug("processRrdpSnapshot():syncLogId:", syncLogId, "notificationModel.Snapshot.Uri, snapshotModel:",
		notificationModel.Snapshot.Uri)
	// first to get snapshot files, because this may fail easily
	snapshotModel, err := getRrdpSnapshot(notificationModel)
	if err != nil {
		belogs.Error("processRrdpSnapshot(): getRrdpSnapshot fail, Snapshot url: ",
			notificationModel.Snapshot.Uri, err)
		return err
	}
	belogs.Info("processRrdpSnapshot():notificationModel.Snapshot.Uri, snapshotModel:",
		notificationModel.Snapshot.Uri,
		snapshotModel.Serial, len(snapshotModel.SnapshotPublishs))

	// rm disk files
	repoHostPath, err := urlutil.JoinPrefixPathAndUrlHost(snapshotDeltaResult.DestPath, notificationModel.Snapshot.Uri)
	if err != nil {
		belogs.Error("processRrdpSnapshot(): JoinPrefixPathAndUrlHost fail, Snapshot url: ",
			notificationModel.Snapshot.Uri, err)
		return err
	}
	snapshotDeltaResult.RepoHostPath = repoHostPath
	belogs.Debug("processRrdpSnapshot():repoHostPath:", repoHostPath)

	err = os.RemoveAll(repoHostPath)
	if err != nil {
		belogs.Error("processRrdpSnapshot(): RemoveAll, repoHostPath: ", repoHostPath, err)
	}
	err = os.MkdirAll(repoHostPath, os.ModePerm)
	if err != nil {
		belogs.Error("processRrdpSnapshot(): MkdirAll, repoHostPath: ", repoHostPath, err)
	}

	// download snapshot files
	rrdpFiles, err := rrdputil.SaveRrdpSnapshotToRrdpFiles(&snapshotModel, snapshotDeltaResult.DestPath)
	if err != nil {
		belogs.Error("processRrdpSnapshot(): SaveRrdpSnapshotToRrdpFiles fail, Snapshot url,  DestPath: ",
			notificationModel.Snapshot.Uri, snapshotDeltaResult.DestPath, err)
		return err
	}
	snapshotDeltaResult.RrdpFiles = rrdpFiles
	belogs.Debug("processRrdpSnapshot():SaveRrdpSnapshotToFiles, notificationModel.Snapshot.Uri, rrdpFiles,snapshotDeltaResult.DestPath:",
		notificationModel.Snapshot.Uri, rrdpFiles, snapshotDeltaResult.DestPath)
	belogs.Info("processRrdpSnapshot():SaveRrdpSnapshotToFiles, notificationModel.Snapshot.Uri, len(rrdpFiles),snapshotDeltaResult.DestPath:",
		notificationModel.Snapshot.Uri, len(rrdpFiles), snapshotDeltaResult.DestPath)

	// del old cer/crl/mft/roa and update to rrdplog
	err = UpdateRrdpSnapshot(syncLogId, notificationModel, &snapshotModel,
		snapshotDeltaResult, syncLogFilesCh)
	if err != nil {
		belogs.Error("processRrdpSnapshot(): UpdateRrdpSnapshot fail, syncLogId, snapshotDeltaResult: ",
			syncLogId, jsonutil.MarshalJson(snapshotDeltaResult), err)
		return err
	}
	return nil

}
