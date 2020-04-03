package rrdp

import (
	belogs "github.com/astaxie/beego/logs"
	rrdputil "github.com/cpusoft/goutil/rrdputil"
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
		belogs.Debug("getRrdpDelta(): delta.Uri, deltaModel.Serial:", notificationModel.Deltas[i].Uri, deltaModel.Serial)
		deltaModels = append(deltaModels, deltaModel)
	}

	return deltaModels, nil
}
