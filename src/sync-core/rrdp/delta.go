package rrdp

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/rrdputil"
)

// lastSerial is last syncRrdpLog's curSerial
func processRrdpDelta(syncLogId uint64, notificationModel *rrdputil.NotificationModel,
	snapshotDeltaResult *SnapshotDeltaResult, syncLogFilesCh chan []model.LabRpkiSyncLogFile) (err error) {

	start := time.Now()
	timeoutMins := conf.Int("rrdp::rrdpDeltaTimeoutSeconds") / 60
	externalVerifyHttps := conf.Bool("sync::externalVerifyHttps")
	httpConfig := httpclient.NewHttpClientConfigWithParam(uint64(timeoutMins), 3, "all", externalVerifyHttps)
	deltaModels, err := rrdputil.GetRrdpDeltasWithConfig(notificationModel, snapshotDeltaResult.LastSerial, httpConfig)
	if err != nil {
		belogs.Error("processRrdpDelta(): GetRrdpDeltasWithConfig fail, notifyUrl:", snapshotDeltaResult.NotifyUrl,
			", len(notificationModel.Deltas): ", len(notificationModel.Deltas), err)
		return err
	}
	belogs.Info("processRrdpDelta():GetRrdpDeltasWithConfig  notifyUrl:", snapshotDeltaResult.NotifyUrl,
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

	snapshotDeltaResult.RrdpFiles = rrdpFilesAll
	belogs.Debug("processRrdpDelta(): notifyUrl:", snapshotDeltaResult.NotifyUrl, "   notificationModel.Snapshot.Uri, snapshotDeltaResult.RrdpFiles, snapshotDeltaResult.DestPath:",
		notificationModel.Snapshot.Uri, jsonutil.MarshalJson(snapshotDeltaResult.RrdpFiles),
		snapshotDeltaResult.DestPath, "   time(s):", time.Since(start))

	// del old cer/crl/mft/roa and update to rrdplog
	// get dest path : /home/rpki/data/reporrdp/
	err = updateRrdpDeltaDb(syncLogId, deltaModels, snapshotDeltaResult, syncLogFilesCh)
	if err != nil {
		belogs.Error("processRrdpDelta(): updateRrdpDeltaDb fail,notifyUrl:", snapshotDeltaResult.NotifyUrl,
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
