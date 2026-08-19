package parsevalidate

import (
	"sync"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	parsevalidatecore "github.com/bgpsecurity/rpstir2/parsevalidate-core"
	parsevalidatedb "github.com/bgpsecurity/rpstir2/parsevalidate-db"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/conf"
	"github.com/cpusoft/goutil/osutil"
)

// ParseValidateStart: start
func parseValidateStart() (nextStep string, err error) {

	start := time.Now()
	belogs.Info("parseValidateStart(): start")
	// save starttime to lab_rpki_sync_log
	labRpkiSyncLogId, err := parsevalidatedb.UpdateSyncLogParseValidateStartDb("parsevalidating")
	if err != nil {
		belogs.Error("parseValidateStart():UpdateSyncLogParseValidateStartDb fail:", err)
		return "", err
	}
	belogs.Debug("parseValidateStart():UpdateSyncLogParseValidateStartDb, labRpkiSyncLogId:", labRpkiSyncLogId, "  time(s):", time.Since(start))

	syncLogFileModelsCh := make(chan []*model.SyncLogFileModel, conf.Int("parse::parseConcurrentCount"))
	syncLogFileModelsDbCh := make(chan []*model.SyncLogFileModel, conf.Int("parse::updateConcurrentCount"))

	var wg sync.WaitGroup
	go callParseValidate(syncLogFileModelsCh, syncLogFileModelsDbCh)
	go callDelAndInsertCert(syncLogFileModelsDbCh, &wg)

	// get rsyncLogFile
	err = parsevalidatedb.GetSyncLogFileModelBySyncLogIdDb(labRpkiSyncLogId, syncLogFileModelsCh, &wg)
	if err != nil {
		belogs.Error("parseValidateStart():GetSyncLogFileModelBySyncLogIdDb fail:", labRpkiSyncLogId, err)
		close(syncLogFileModelsCh)
		close(syncLogFileModelsDbCh)
		return "", err
	}
	belogs.Debug("parseValidateStart(): GetSyncLogFileModelBySyncLogIdDb, labRpkiSyncLogId:", labRpkiSyncLogId, "  will wait,  time(s):", time.Since(start))

	wg.Wait()
	close(syncLogFileModelsCh)
	close(syncLogFileModelsDbCh)
	belogs.Info("parseValidateStart(): all done, labRpkiSyncLogId:", labRpkiSyncLogId, "   time(s):", time.Since(start))

	// save to db
	err = parsevalidatedb.UpdateSyncLogParseValidateStateEndDb(labRpkiSyncLogId, "parsevalidated", make([]string, 0))
	if err != nil {
		belogs.Error("parseValidateStart(): UpdateSyncLogParseValidateStateEndDb fail: ", err)
		return "", err
	}

	belogs.Info("parseValidateStart(): end, will call chainvalidate,  time(s):", time.Since(start))
	return "chainvalidate", nil

}
func callParseValidate(syncLogFileModelsCh, syncLogFileModelsDbCh chan []*model.SyncLogFileModel) {
	for {
		select {
		case syncLogFileModels, ok := <-syncLogFileModelsCh:
			belogs.Debug("callParseValidate(): len(syncLogFileModels)", len(syncLogFileModels), "  ok:", ok)
			if ok {
				go parseValidate(syncLogFileModels, syncLogFileModelsDbCh)
			} else {
				belogs.Debug("callParseValidate(): closed, will return")
				return
			}
		}
	}
}

func parseValidate(syncLogFileModels []*model.SyncLogFileModel, syncLogFileModelsDbCh chan []*model.SyncLogFileModel) {
	start := time.Now()
	belogs.Debug("parseValidate(): len(syncLogFileModels):", len(syncLogFileModels))
	var err error
	syncLogFileModelsValidated := make([]*model.SyncLogFileModel, 0, len(syncLogFileModels)+10)
	for i, _ := range syncLogFileModels {
		startOne := time.Now()
		//syncLogFileModel := syncLogFileModels[i]
		belogs.Debug("parseValidate(): will parseValidateImpl, id:", syncLogFileModels[i].Id, "  filePathName:", syncLogFileModels[i].FilePath, syncLogFileModels[i].FileName,
			"  fileType:", syncLogFileModels[i].FileType, "   syncType:", syncLogFileModels[i].SyncType)
		if syncLogFileModels[i].SyncType == "update" || syncLogFileModels[i].SyncType == "add" {
			// process "add" and "update" rsyncLogFile
			belogs.Debug("parseValidate(): will parseValidateImpl, id:", syncLogFileModels[i].Id, "  filePathName:", syncLogFileModels[i].FilePath, syncLogFileModels[i].FileName,
				"  fileType:", syncLogFileModels[i].FileType, "   syncType:", syncLogFileModels[i].SyncType)
			err = parseValidateImpl(syncLogFileModels[i])
			if err != nil {
				belogs.Error("parseValidate(): parseValidateImpl fail, syncLogFileModel:", syncLogFileModels[i].String(),
					err, "  time(s):", time.Since(startOne))
				continue
			}
		}
		belogs.Debug("parseValidate(): add to syncLogFileModelsValidated, will send, id:", syncLogFileModels[i].Id, "  filePathName:", syncLogFileModels[i].FilePath, syncLogFileModels[i].FileName,
			"  fileType:", syncLogFileModels[i].FileType, "   syncType:", syncLogFileModels[i].SyncType, "  time(s):", time.Since(startOne))
		syncLogFileModelsValidated = append(syncLogFileModelsValidated, syncLogFileModels[i])
	}
	syncLogFileModelsDbCh <- syncLogFileModelsValidated
	belogs.Info("parseValidate(): send this group syncLogFileModels to update db, len(syncLogFileModelsValidated):", len(syncLogFileModelsValidated),
		"   len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Since(start))
}

func parseValidateImpl(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()

	belogs.Debug("parseValidateImpl(): syncLogFileModel:", syncLogFileModel.String())
	filePathName := osutil.JoinPathFile(syncLogFileModel.FilePath, syncLogFileModel.FileName)
	_, certModel, stateModel, originModel, _, err := parsevalidatecore.ParseValidateFile(filePathName)
	if err != nil {
		belogs.Error("parseValidateImpl(): ParseValidateFile fail, filePathName:", filePathName, err)
		return err
	}
	syncLogFileModel.CertModel = certModel
	syncLogFileModel.OriginModel = originModel
	syncLogFileModel.StateModel = stateModel
	belogs.Debug("parseValidateImpl(): ParseValidateFile ok, filePathName:", filePathName,
		"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType, "  time(s):", time.Since(start))
	return nil
}

func callDelAndInsertCert(syncLogFileModelsDbCh chan []*model.SyncLogFileModel, wg *sync.WaitGroup) {
	for {
		select {
		case syncLogFileModels, ok := <-syncLogFileModelsDbCh:
			belogs.Debug("callDelAndInsertCert(): len(syncLogFileModels):", len(syncLogFileModels), "  ok:", ok)
			if ok {
				go delAndInsertCert(syncLogFileModels, wg)
			} else {
				belogs.Debug("callDelAndInsertCert(): closed, will  return")
				return
			}
		}
	}

}

func delAndInsertCert(syncLogFileModels []*model.SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		belogs.Debug("delAndInsertCert(): wg.Done, len(syncLogFileModels):", len(syncLogFileModels))
		wg.Done()
	}()

	start := time.Now()
	for i, _ := range syncLogFileModels {
		startOne := time.Now()
		syncLogFileModel := syncLogFileModels[i]
		syncType := syncLogFileModel.SyncType
		fileType := syncLogFileModel.FileType
		belogs.Debug("delAndInsertCert(): filePathName:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
			"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType)
		switch fileType {
		case "cer":
			err = parsevalidatedb.DelCerDb(syncLogFileModel)
			if err == nil && (syncType == "update" || syncType == "add") {
				err = parsevalidatedb.InsertCerDb(syncLogFileModel)
			}
		case "crl":
			err = parsevalidatedb.DelCrlDb(syncLogFileModel)
			if err == nil && (syncType == "update" || syncType == "add") {
				err = parsevalidatedb.InsertCrlDb(syncLogFileModel)
			}
		case "mft":
			err = parsevalidatedb.DelMftDb(syncLogFileModel)
			if err == nil && (syncType == "update" || syncType == "add") {
				err = parsevalidatedb.InsertMftDb(syncLogFileModel)
			}
		case "roa":
			err = parsevalidatedb.DelRoaDb(syncLogFileModel)
			if err == nil && (syncType == "update" || syncType == "add") {
				err = parsevalidatedb.InsertRoaDb(syncLogFileModel)
			}
		case "asa":
			err = parsevalidatedb.DelAsaDb(syncLogFileModel)
			if err == nil && (syncType == "update" || syncType == "add") {
				err = parsevalidatedb.InsertAsaDb(syncLogFileModel)
			}
		case "moa":
			err = parsevalidatedb.DelMoaDb(syncLogFileModel)
			if err == nil && (syncType == "update" || syncType == "add") {
				err = parsevalidatedb.InsertMoaDb(syncLogFileModel)
			}
		}
		if err != nil {
			belogs.Error("delAndInsertCert(): Del or Insert fail, syncLogFileModel:", syncLogFileModel.String(), err, " time(s):", time.Since(startOne))
			//return err
			continue // go next
		}
		belogs.Info("delAndInsertCert(): Del or Insert ok, filePathName:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
			"  fileType:", syncLogFileModel.FileType, "   syncType:", syncLogFileModel.SyncType, " time(s):", time.Since(startOne))
	}
	belogs.Info("delAndInsertCert(): save all, len(syncLogFileModels):", len(syncLogFileModels), " time(s):", time.Since(start))

	return nil
}
