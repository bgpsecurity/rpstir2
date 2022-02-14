package rrdp

import (
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/rrdputil"
)

func GetRrdpNotification(notifyUrl string) (notificationModel rrdputil.NotificationModel, err error) {

	belogs.Debug("GetRrdpNotification(): notifyUrl :", notifyUrl)
	// get nofiy.xmlm
	notificationModel, err = rrdputil.GetRrdpNotification(notifyUrl)
	if err != nil {
		belogs.Error("GetRrdpNotification(): GetRrdpNotification fail, notifyUrl: ",
			notifyUrl, err)
		return notificationModel, err
	}
	err = rrdputil.CheckRrdpNotification(&notificationModel)
	if err != nil {
		belogs.Error("GetRrdpNotification(): CheckRrdpNotification fail, notifyUrl,notificationModel: ",
			notifyUrl, jsonutil.MarshalJson(notificationModel), err)
		return notificationModel, err
	}
	return notificationModel, nil
}
