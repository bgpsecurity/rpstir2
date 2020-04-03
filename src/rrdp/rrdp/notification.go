package rrdp

import (
	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	rrdputil "github.com/cpusoft/goutil/rrdputil"
)

func getRrdpNotification(notifyUrl string) (notificationModel rrdputil.NotificationModel, err error) {

	belogs.Debug("getRrdpNotification(): notifyUrl :", notifyUrl)
	// get nofiy.xmlm
	notificationModel, err = rrdputil.GetRrdpNotification(notifyUrl)
	if err != nil {
		belogs.Error("getRrdpNotification(): GetRrdpNotification fail, notifyUrl: ",
			notifyUrl, err)
		return notificationModel, err
	}
	err = rrdputil.CheckRrdpNotification(&notificationModel)
	if err != nil {
		belogs.Error("getRrdpNotification(): CheckRrdpNotification fail, notifyUrl,notificationModel: ",
			notifyUrl, jsonutil.MarshalJson(notificationModel), err)
		return notificationModel, err
	}
	return notificationModel, nil
}
