package rrdp

import (
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/rrdputil"
)

// lastSerial is last syncRrdpLog's curSerial
func processRrdpDelta(syncLogId uint64, notificationModel *rrdputil.NotificationModel,
	snapshotDeltaResult *SnapshotDeltaResult, syncLogFilesCh chan []model.SyncLogFile) (err error) {

	start := time.Now()
	deltaModels, err := rrdputil.GetRrdpDeltas(notificationModel, snapshotDeltaResult.LastSerial)
	if err != nil {
		belogs.Error("processRrdpDelta(): GetRrdpDeltas fail, notifyUrl:", snapshotDeltaResult.NotifyUrl,
			", len(notificationModel.Deltas): ", len(notificationModel.Deltas), err)
		return err
	}
	belogs.Info("processRrdpDelta():GetRrdpDeltas  notifyUrl:", snapshotDeltaResult.NotifyUrl,
		"   len(deltaModels):", len(deltaModels))
	if len(deltaModels) <= 0 {
		belogs.Debug("processRrdpDelta():notifyUrl:", snapshotDeltaResult.NotifyUrl, "   len(deltaModels)<=0:", len(deltaModels))
		return nil
	}

	rrdpFilesAll, err := rrdputil.SaveRrdpDeltasToRrdpFiles(deltaModels, snapshotDeltaResult.NotifyUrl, snapshotDeltaResult.DestPath)
	if err != nil {
		belogs.Error("processRrdpDelta(): SaveRrdpDeltasToRrdpFiles fail, notifyUrl:", snapshotDeltaResult.NotifyUrl,
			"   len(deltaModels):", len(deltaModels),
			"   snapshotDeltaResult.DestPath: ", snapshotDeltaResult.DestPath, err)
		return err
	}
	/*
		rrdpFilesAll := make([]rrdputil.RrdpFile, 0)
		// download snapshot files
		for i := range deltaModels {
			// save publish files and remove withdraw files
			rrdpFiles, err := rrdputil.SaveRrdpDeltaToRrdpFiles(&deltaModels[i], snapshotDeltaResult.DestPath)
			if err != nil {
				belogs.Error("processRrdpDelta(): SaveRrdpDeltaToRrdpFiles fail, notifyUrl:", snapshotDeltaResult.NotifyUrl,
					"   deltaModels[i].SessionId:", deltaModels[i].SessionId,
					"   deltaModels[i].Serial:", deltaModels[i].Serial, "   deltaModels[i].DeltaUrl:", deltaModels[i].DeltaUrl,
					"   snapshotDeltaResult.DestPath: ", snapshotDeltaResult.DestPath, err)
				return err
			}
			rrdpFilesAll = append(rrdpFilesAll, rrdpFiles...)
		}
	*/
	snapshotDeltaResult.RrdpFiles = rrdpFilesAll
	belogs.Debug("processRrdpDelta(): notifyUrl:", snapshotDeltaResult.NotifyUrl, "   notificationModel.Snapshot.Uri, snapshotDeltaResult.RrdpFiles, snapshotDeltaResult.DestPath:",
		notificationModel.Snapshot.Uri, jsonutil.MarshalJson(snapshotDeltaResult.RrdpFiles),
		snapshotDeltaResult.DestPath, "   time(s):", time.Since(start))

	// del old cer/crl/mft/roa and update to rrdplog
	// get dest path : /root/rpki/data/reporrdp/
	err = UpdateRrdpDelta(syncLogId, deltaModels, snapshotDeltaResult, syncLogFilesCh)
	if err != nil {
		belogs.Error("processRrdpDelta(): UpdateRrdpDelta fail,notifyUrl:", snapshotDeltaResult.NotifyUrl,
			"    Snapshot url:", notificationModel.Snapshot.Uri,
			"    repoPath: ", snapshotDeltaResult.DestPath, err, "   time(s):", time.Since(start))
		return err
	}
	belogs.Info("processRrdpDelta(): notifyUrl:", snapshotDeltaResult.NotifyUrl,
		"     Snapshot.Uri:", notificationModel.Snapshot.Uri,
		"     len(rrdpFiles):", len(snapshotDeltaResult.RrdpFiles),
		"     snapshotDeltaResult.DestPath:", snapshotDeltaResult.DestPath, "   time(s):", time.Since(start))

	return nil
}
