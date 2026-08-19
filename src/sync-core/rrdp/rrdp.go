package rrdp

import (
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/hashutil"
	"github.com/cpusoft/goutil/httpclient"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
	"github.com/cpusoft/goutil/rrdputil"
)

// connectRrdpUrlCh: whether connect to notifyurl, will tell others to remove rsync path, or just ignore. will defer close(connectRrdpUrlCh)
// rrdpFiles: rrdp results files, if len() is 0, just rrdp is no update
func RrdpByUrlImpl(rrdpByUrlModel RrdpByUrlModel, connectRrdpUrlCh chan bool,
	syncLogFilesCh chan []model.LabRpkiSyncLogFile) (rrdpFiles []rrdputil.RrdpFile, err error) {
	start := time.Now()
	//defer RrdpByUrlDefer(rrdpUrl, rrdpUrlCh)
	belogs.Debug("RrdpByUrlImpl():start, rrdpByUrlModel:", jsonutil.MarshalJson(rrdpByUrlModel))
	timeoutMins := conf.Int("rrdp::rrdpNoticationTimeoutSeconds") / 60
	externalVerifyHttps := conf.Bool("sync::externalVerifyHttps")
	// get notify xml
	notificationModel, err := rrdputil.GetAndCheckRrdpNotificationWithConfig(rrdpByUrlModel.NotifyUrl,
		httpclient.NewHttpClientConfigWithParam(uint64(timeoutMins), 1, "all", externalVerifyHttps))
	if err != nil {
		// connect false
		connectRrdpUrlCh <- false
		close(connectRrdpUrlCh)
		belogs.Error("RrdpByUrlImpl(): GetAndCheckRrdpNotificationWithConfig fail, rrdpByUrlModel:",
			jsonutil.MarshalJson(rrdpByUrlModel), "  will send false to connectRrdpUrlCh,  err:", err,
			"  time(s):", time.Since(start))
		return nil, err
	}

	connectRrdpUrlCh <- true
	close(connectRrdpUrlCh)
	// connect ok
	belogs.Info("RrdpByUrlImpl(): connect Notify Url ok:", rrdpByUrlModel.NotifyUrl,
		", will send true to connectRrdpUrlCh ,  time(s):", time.Since(start))

	// no need update
	belogs.Debug("RrdpByUrlImpl(): compare :",
		"   rrdpByUrlModel:", jsonutil.MarshalJson(rrdpByUrlModel),
		"   notificationModel.SessionId:", notificationModel.SessionId,
		"   notificationModel.MaxSerial:", notificationModel.MaxSerial)
	if rrdpByUrlModel.HasPath && rrdpByUrlModel.HasLast &&
		rrdpByUrlModel.LastSessionId == notificationModel.SessionId &&
		rrdpByUrlModel.LastCurSerial == notificationModel.MaxSerial {
		belogs.Info("RrdpByUrlImpl(): no new rrdp serial, no need rrdp to download, just return:", rrdpByUrlModel.NotifyUrl)
		return nil, nil
	}

	// check delta or snapshot
	canDelta := false
	if rrdpByUrlModel.HasPath && rrdpByUrlModel.HasLast &&
		rrdpByUrlModel.LastSessionId == notificationModel.SessionId &&
		rrdpByUrlModel.LastCurSerial >= notificationModel.MinSerial &&
		rrdpByUrlModel.LastCurSerial < notificationModel.MaxSerial {
		canDelta = true
	}
	belogs.Info("RrdpByUrlImpl():notifyUrl canDelta:", rrdpByUrlModel.NotifyUrl, canDelta)

	// need to get snapshot
	var snapshotDeltaResult SnapshotDeltaResult
	if !canDelta {
		snapshotDeltaResult = SnapshotDeltaResult{
			NotifyUrl:  rrdpByUrlModel.NotifyUrl,
			DestPath:   rrdpByUrlModel.DestPath,
			LastSerial: 0}
		belogs.Info("RrdpByUrlImpl(): will snapshot:", rrdpByUrlModel.NotifyUrl, jsonutil.MarshalJson(snapshotDeltaResult))
		err = processRrdpSnapshot(rrdpByUrlModel.SyncLogId, &notificationModel,
			&snapshotDeltaResult, syncLogFilesCh)

	} else {
		// get delta
		snapshotDeltaResult = SnapshotDeltaResult{
			NotifyUrl:  rrdpByUrlModel.NotifyUrl,
			DestPath:   rrdpByUrlModel.DestPath,
			LastSerial: rrdpByUrlModel.LastCurSerial}
		belogs.Info("RrdpByUrlImpl(): will delta:", rrdpByUrlModel.NotifyUrl, jsonutil.MarshalJson(snapshotDeltaResult))
		err = processRrdpDelta(rrdpByUrlModel.SyncLogId, &notificationModel,
			&snapshotDeltaResult, syncLogFilesCh)
	}

	if err != nil {
		belogs.Error("RrdpByUrlImpl(): processRrdpSnapshot or  processRrdpDelta fail, canDelta:",
			canDelta, jsonutil.MarshalJson(rrdpByUrlModel), err)
		return nil, err
	}
	belogs.Info("RrdpByUrlImpl(): end ok, notifyUrl, len(files):", rrdpByUrlModel.NotifyUrl, len(snapshotDeltaResult.RrdpFiles),
		"  time(s):", time.Since(start))
	return snapshotDeltaResult.RrdpFiles, nil

}

// syncType:add/update/del
// syncStyle: rrdp/rsync
func ConvertToSyncLogFile(
	syncLogId uint64, rrdpTime time.Time,
	rrdpFile *rrdputil.RrdpFile) (syncLogFile model.LabRpkiSyncLogFile, err error) {
	belogs.Debug("ConvertToSyncLogFile():rrdp, syncLogId, rrdpTime, rrdpFile:",
		syncLogId, rrdpTime, jsonutil.MarshalJson(rrdpFile))

	// filehash
	var fileHash string
	if rrdpFile.SyncType == "add" || rrdpFile.SyncType == "update" {
		file := osutil.JoinPathFile(rrdpFile.FilePath, rrdpFile.FileName)
		fileHash, err = hashutil.Sha256File(file)
		if err != nil {
			belogs.Error("ConvertToSyncLogFile():Sha256File fail, file:", file, jsonutil.MarshalJson(rrdpFile), err)
			return syncLogFile, err
		}
	}

	// filetype
	fileType := osutil.ExtNoDot(rrdpFile.FileName)
	rtr := "notNeed"
	if fileType == "roa" || fileType == "asa" || fileType == "moa" {
		rtr = "notYet"
	}

	syncLogFileState := model.LabRpkiSyncLogFileState{
		Sync:            "finished",
		UpdateCertTable: "notYet",
		Rtr:             rtr,
	}
	state := jsonutil.MarshalJson(syncLogFileState)

	// syncLogFile
	syncLogFile = model.LabRpkiSyncLogFile{
		SyncLogId: syncLogId,
		FileType:  fileType,
		SyncTime:  rrdpTime,
		FilePath:  rrdpFile.FilePath,
		FileName:  rrdpFile.FileName,
		SourceUrl: rrdpFile.SourceUrl,
		SyncType:  rrdpFile.SyncType,
		SyncStyle: "rrdp",
		State:     state,
		FileHash:  fileHash,
	}
	return syncLogFile, nil
}

var duplicatedRrdpUrls map[string]int

// duplicated snapshot:
// https://chloe.sobornost.net/rpki/news.xml and https://chloe.sobornost.net/rpki/news-public.xml
// https://rpki.co/rrdp/notification.xml and https://rrdp.rpki.co/rrdp/notification.xml
func init() {
	duplicatedRrdpUrls = make(map[string]int)
	duplicatedRrdpUrls[`https://chloe.sobornost.net/rpki/news-public.xml`] = 1
	duplicatedRrdpUrls[`https://rpki.co/rrdp/notification.xml`] = 1
}

func IsDuplicatedRrdpUrls(rrdpUrl string) bool {
	belogs.Debug("IsDuplicatedRrdpUrls(): rrdpUrl:", rrdpUrl)
	_, ok := duplicatedRrdpUrls[rrdpUrl]
	return ok
}
