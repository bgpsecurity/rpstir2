package rrdp

import (
	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	rrdputil "github.com/cpusoft/goutil/rrdputil"
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
			notificationModel.Snapshot.Uri, jsonutil.MarshalJson(snapshotModel), err)
		return snapshotModel, err
	}
	return snapshotModel, nil

}
