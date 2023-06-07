package parsevalidate

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/osutil"
)

var parseCount int64

// ParseValidateStart: start
func parseValidateStart2() (nextStep string, err error) {

	start := time.Now()
	parseCount = 0
	belogs.Info("parseValidateStart(): start")
	// save starttime to lab_rpki_sync_log
	labRpkiSyncLogId, err := updateRsyncLogParseValidateStartDb("parsevalidating")
	if err != nil {
		belogs.Error("parseValidateStart():updateRsyncLogParseValidateStartDb fail:", err)
		return "", err
	}
	belogs.Debug("parseValidateStart():updateRsyncLogParseValidateStartDb, labRpkiSyncLogId:", labRpkiSyncLogId, "  time(s):", time.Since(start))

	parseConcurrentCh := make(chan int, conf.Int("parse::parseConcurrentCount"))
	syncLogFileModelCh := make(chan *SyncLogFileModel, runtime.NumCPU()*2)
	endCh := make(chan bool, 1)
	var wg sync.WaitGroup
	go callParseValidate(syncLogFileModelCh, parseConcurrentCh, endCh, &wg)
	defer func() {
		endCh <- true
		close(parseConcurrentCh)
		close(syncLogFileModelCh)
		close(endCh)
	}()
	// get rsyncLogFile
	err = getSyncLogFileModelBySyncLogIdDb(labRpkiSyncLogId, syncLogFileModelCh, parseConcurrentCh, endCh, &wg)
	if err != nil {
		belogs.Error("parseValidateStart():getSyncLogFileModelBySyncLogIdDb fail:", labRpkiSyncLogId, err)
		return "", err
	}
	belogs.Debug("parseValidateStart(): getSyncLogFileModelBySyncLogIdDb, labRpkiSyncLogId:", labRpkiSyncLogId)

	wg.Wait()
	belogs.Info("parseValidateStart(): all done, labRpkiSyncLogId:", labRpkiSyncLogId, "   time(s):", time.Since(start))

	// save to db
	err = updateRsyncLogParseValidateStateEndDb(labRpkiSyncLogId, "parsevalidated", make([]string, 0))
	if err != nil {
		belogs.Error("parseValidateStart(): updateRsyncLogParseValidateStateEndDb fail: ", err)
		return "", err
	}

	belogs.Info("parseValidateStart(): end, will call chainvalidate,  time(s):", time.Since(start))
	return "chainvalidate", nil

}
func callParseValidate(syncLogFileModelCh chan *SyncLogFileModel, parseConcurrentCh chan int, endCh chan bool, wg *sync.WaitGroup) {
	for {
		select {
		case syncLogFileModel, ok := <-syncLogFileModelCh:
			belogs.Debug("callParseValidate(): id:", syncLogFileModel.Id, "  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
				"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType, ok)
			if ok && syncLogFileModel.Id > 0 {
				go parseValidate(syncLogFileModel, parseConcurrentCh, wg)
			}

		case end := <-endCh:
			if end {
				belogs.Debug("callParseValidate(): end:")
				return
			}
		}
	}
}

//
func parseValidate(syncLogFileModel *SyncLogFileModel, parseConcurrentCh chan int, wg *sync.WaitGroup) {
	defer func() {
		belogs.Debug("parseValidate(): wg.Done, id:", syncLogFileModel.Id,
			"  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
			"  index:", syncLogFileModel.Index)
		atomic.AddInt64(&parseCount, -1)
		wg.Done()
		// have done this, so will allow get new syncLogFileModel from db
		<-parseConcurrentCh
		belogs.Debug("parseValidate(): AddInt64(-1), parseCount:", atomic.LoadInt64(&parseCount))
	}()

	start := time.Now()
	belogs.Debug("parseValidate(): id:", syncLogFileModel.Id, "  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType)

	var err error
	// add/del/update, need del first
	belogs.Debug("parseValidate(): will delCert, id:", syncLogFileModel.Id, "  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType)
	err = delCert(syncLogFileModel)
	if err != nil {
		belogs.Error("parseValidate(): delCert() fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel),
			err, "  time(s):", time.Since(start))
		return
	}
	belogs.Debug("parseValidate(): delCert, id:", syncLogFileModel.Id, "  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType, "  time(s):", time.Since(start))

	// only update/add will parseValidate and insert
	if syncLogFileModel.SyncType == "update" || syncLogFileModel.SyncType == "add" {
		// process "add" and "update" rsyncLogFile
		belogs.Debug("parseValidate(): will parseValidateAndInsertCert, id:", syncLogFileModel.Id, "  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
			"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType)
		err = parseValidateAndInsertCert(syncLogFileModel)
		if err != nil {
			belogs.Error("parseValidateStart():parseValidateAndInsertCert fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel),
				err, "  time(s):", time.Since(start))
			return
		}
	}
	belogs.Info("parseValidate(): ok, del or update or insert id:", syncLogFileModel.Id,
		"  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType,
		"  time(s):", time.Since(start))
}

// get del;
// when update, because "update" should del first
func parseValidateAndInsertCert(syncLogFileModel *SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("parseValidateAndInsertCert(): id:", syncLogFileModel.Id,
		"  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType)

	belogs.Debug("parseValidateAndInsertCert(): syncLogFileModel :", jsonutil.MarshalJson(syncLogFileModel))
	file := osutil.JoinPathFile(syncLogFileModel.FilePath, syncLogFileModel.FileName)
	belogs.Debug("parseValidateAndInsertCert(): file :", file)
	_, certModel, stateModel, err := parseValidateFile(file)
	if err != nil {
		belogs.Error("parseValidateAndInsertCert(): parseValidateFile fail: ", file, err)
		return err
	}
	syncLogFileModel.CertModel = certModel
	syncLogFileModel.StateModel = stateModel
	belogs.Debug("parseValidateAndInsertCert(): parseValidateFile file :", file,
		"   syncType:", syncLogFileModel.SyncType, "  time(s):", time.Since(start))

	switch syncLogFileModel.FileType {
	case "cer":
		err = addCerDb(syncLogFileModel)
	case "crl":
		err = addCrlDb(syncLogFileModel)
	case "mft":
		err = addMftDb(syncLogFileModel)
	case "roa":
		err = addRoaDb(syncLogFileModel)
	case "asa":
		err = addAsaDb(syncLogFileModel)
	}
	if err != nil {
		belogs.Error("parseValidateAndInsertCert(): add***Db fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err, " time(s):", time.Since(start))
		return err
	}
	belogs.Debug("parseValidateAndInsertCert(): parseAndValidate ok, id:", syncLogFileModel.Id,
		"  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType, " time(s):", time.Since(start))
	return nil

}

// get insert;
func delCert(syncLogFileModel *SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("delCert(): id:", syncLogFileModel.Id,
		"  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType)

	switch syncLogFileModel.FileType {
	case "cer":
		err = delCerDb(syncLogFileModel)
	case "crl":
		err = delCrlDb(syncLogFileModel)
	case "mft":
		err = delMftDb(syncLogFileModel)
	case "roa":
		err = delRoaDb(syncLogFileModel)
	case "asa":
		err = delAsaDb(syncLogFileModel)
	}
	if err != nil {
		belogs.Error("delCert(): del***Db fail, syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel), err, " time(s):", time.Since(start))
		return err
	}
	belogs.Info("delCert(): del ok, id:", syncLogFileModel.Id,
		"  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType, " time(s):", time.Since(start))
	return nil

}
