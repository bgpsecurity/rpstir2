package rrdp

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/cpusoft/goutil/jsonutil"
	rrdputil "github.com/cpusoft/goutil/rrdputil"

	"rrdp/db"
	rrdpmodel "rrdp/model"
)

func getRrdpDelta(notificationModel *rrdputil.NotificationModel, lastSerial uint64) (deltaModels []rrdputil.DeltaModel, err error) {

	belogs.Debug("getRrdpDelta(): len(notificationModel.MapSerialDeltas),lastSerial :",
		len(notificationModel.MapSerialDeltas), lastSerial)

	deltaModels = make([]rrdputil.DeltaModel, 0, len(notificationModel.MapSerialDeltas))
	// serial need from small to large
	for i := len(notificationModel.Deltas) - 1; i >= 0; i-- {

		belogs.Debug("getRrdpDelta():notificationModel.Deltas[i].Serial:", notificationModel.Deltas[i].Serial)
		if notificationModel.Deltas[i].Serial <= lastSerial {
			belogs.Debug("getRrdpDelta():continue, notificationModel.Deltas[i].Serial <= lastSerial:", notificationModel.Deltas[i].Serial, lastSerial)
			continue
		}

		deltaModel, err := rrdputil.GetRrdpDelta(notificationModel.Deltas[i].Uri)
		if err != nil {
			belogs.Error("getRrdpDelta(): GetRrdpDelta fail, delta.Uri :",
				notificationModel.Deltas[i].Uri, err)
			return deltaModels, err
		}

		err = rrdputil.CheckRrdpDelta(&deltaModel, notificationModel)
		if err != nil {
			belogs.Error("getRrdpDelta(): CheckRrdpDelta fail, delta.Uri :",
				notificationModel.Deltas[i].Uri, err)
			return deltaModels, err
		}
		belogs.Debug("getRrdpDelta(): delta.Uri, deltaModel:", notificationModel.Deltas[i].Uri, jsonutil.MarshalJson(deltaModel))
		deltaModels = append(deltaModels, deltaModel)
	}

	return deltaModels, nil
}

// lastSerial is last syncRrdpLog's curSerial
func processRrdpDelta(syncLogId uint64, notificationModel *rrdputil.NotificationModel,
	snapshotDeltaResult *rrdpmodel.SnapshotDeltaResult) (err error) {

	deltaModels, err := getRrdpDelta(notificationModel, snapshotDeltaResult.LastSerial)
	if err != nil {
		belogs.Error("processRrdpDelta(): getRrdpDelta fail,  len(notificationModel.MapSerialDeltas) :",
			len(notificationModel.MapSerialDeltas), err)
		return err
	}
	belogs.Debug("processRrdpDelta():getRrdpDelta len(deltaModels):", len(deltaModels))
	if len(deltaModels) <= 0 {
		return nil
	}

	rrdpFilesAll := make([]rrdputil.RrdpFile, 0)
	// download snapshot files
	for i := range deltaModels {
		// save publish files and remove withdraw files
		rrdpFiles, err := rrdputil.SaveRrdpDeltaToRrdpFiles(&deltaModels[i], snapshotDeltaResult.DestPath)
		if err != nil {
			belogs.Error("processRrdpDelta(): SaveRrdpDeltaToRrdpFiles fail, deltaModels[i],  snapshotDeltaResult.DestPath: ",
				jsonutil.MarshalJson(deltaModels[i]), snapshotDeltaResult.DestPath, err)
			return err
		}
		rrdpFilesAll = append(rrdpFilesAll, rrdpFiles...)
	}
	snapshotDeltaResult.RrdpFiles = rrdpFilesAll
	belogs.Debug("processRrdpDelta():SaveRrdpDeltaToRrdpFiles notificationModel.Snapshot.Uri, snapshotDeltaResult.RrdpFiles, snapshotDeltaResult.DestPath:",
		notificationModel.Snapshot.Uri, jsonutil.MarshalJson(snapshotDeltaResult.RrdpFiles), snapshotDeltaResult.DestPath)
	belogs.Info("processRrdpDelta():SaveRrdpDeltaToRrdpFiles len(rrdpFiles), snapshotDeltaResult.DestPath:",
		notificationModel.Snapshot.Uri, len(snapshotDeltaResult.RrdpFiles), snapshotDeltaResult.DestPath)

	// del old cer/crl/mft/roa and update to rrdplog
	// get dest path : /root/rpki/data/reporrdp/
	err = db.UpdateRrdpDelta(syncLogId, deltaModels, snapshotDeltaResult)
	if err != nil {
		belogs.Error("ProcessRrdpSnapshot(): SaveRrdpSnapshotToFiles fail, Snapshot url,  repoPath: ",
			notificationModel.Snapshot.Uri, snapshotDeltaResult.DestPath, err)
		return err
	}
	return nil
}
