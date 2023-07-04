package parsevalidate

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

func getSyncLogFileModelsBySyncLogIdDb(labRpkiSyncLogId uint64) (syncLogFileModels *SyncLogFileModels, err error) {
	start := time.Now()

	belogs.Debug("getSyncLogFileModelsBySyncLogIdDb():start")
	dbSyncLogFileModels := make([]SyncLogFileModel, 0)
	sql := `select s.id,s.syncLogId,s.filePath,s.fileName, s.fileType, s.syncType, 
				cast(CONCAT(IFNULL(c.id,''),IFNULL(m.id,''),IFNULL(l.id,''),IFNULL(r.id,''),IFNULL(a.id,'')) as unsigned int) as certId from lab_rpki_sync_log_file s 
			left join lab_rpki_cer c on c.filePath = s.filePath and c.fileName = s.fileName  
			left join lab_rpki_mft m on m.filePath = s.filePath and m.fileName = s.fileName  
			left join lab_rpki_crl l on l.filePath = s.filePath and l.fileName = s.fileName  
			left join lab_rpki_roa r on r.filePath = s.filePath and r.fileName = s.fileName 
			left join lab_rpki_asa a on a.filePath = s.filePath and a.fileName = s.fileName 
			where s.state->>'$.updateCertTable'='notYet' and s.syncLogId=? order by s.id `
	err = xormdb.XormEngine.SQL(sql, labRpkiSyncLogId).Find(&dbSyncLogFileModels)
	if err != nil {
		belogs.Error("getSyncLogFileModelsBySyncLogIdDb(): Find fail:", err)
		return nil, err
	}
	belogs.Debug("getSyncLogFileModelsBySyncLogIdDb(): len(dbSyncLogFileModels):", len(dbSyncLogFileModels), jsonutil.MarshalJson(dbSyncLogFileModels))
	syncLogFileModels = NewSyncLogFileModels(labRpkiSyncLogId, dbSyncLogFileModels)
	belogs.Info("getSyncLogFileModelsBySyncLogIdDb(): end, len(dbSyncLogFileModels),  time(s):", len(dbSyncLogFileModels), time.Since(start))
	return syncLogFileModels, nil
}

func updateSyncLogFilesJsonAllAndStateDb(session *xorm.Session, syncLogFileModels []SyncLogFileModel) error {
	belogs.Debug("updateSyncLogFilesJsonAllAndStateDb(): len(syncLogFileModels):", len(syncLogFileModels))
	sqlStr := `update lab_rpki_sync_log_file f set 	
	  f.state=json_replace(f.state,'$.updateCertTable','finished','$.rtr',?) ,
	  f.jsonAll=?  where f.id=?`
	for i := range syncLogFileModels {
		rtrState := "notNeed"
		jsonAll := ""
		if (syncLogFileModels[i].FileType == "roa" || syncLogFileModels[i].FileType == "asa") &&
			syncLogFileModels[i].SyncType != "del" {
			rtrState = "notYet"
		}

		//when del or update(before del), syncLogFileModels[i].CertModel is nil
		if syncLogFileModels[i].CertModel == nil {
			belogs.Debug("updateSyncLogFilesJsonAllAndStateDb(): del or update, CertModel is nil:",
				jsonutil.MarshalJson(syncLogFileModels[i]))
		} else {
			// when add or update(after del), syncLogFileModels[i].CertModel is not nil
			/*
				switch syncLogFileModels[i].FileType {
				case "cer":
					cerModel := syncLogFileModels[i].CertModel.(model.CerModel)
					jsonAll = jsonutil.MarshalJson(cerModel)
				case "crl":
					crlModel := syncLogFileModels[i].CertModel.(model.CrlModel)
					jsonAll = jsonutil.MarshalJson(crlModel)
				case "mft":
					mftModel := syncLogFileModels[i].CertModel.(model.MftModel)
					jsonAll = jsonutil.MarshalJson(mftModel)
				case "roa":
					roaModel := syncLogFileModels[i].CertModel.(model.RoaModel)
					jsonAll = jsonutil.MarshalJson(roaModel)
				default:
					belogs.Error("updateSyncLogFilesJsonAllAndStateDb(): syncLogFileModels[i].FileType fail:",
						syncLogFileModels[i].FileType)
					return errors.New("syncLogFileModels[i].FileType fail, " + syncLogFileModels[i].FileType)
				}
			*/
			jsonAll = jsonutil.MarshalJson(syncLogFileModels[i].CertModel)
		}

		_, err := session.Exec(sqlStr, rtrState, xormdb.SqlNullString(jsonAll), syncLogFileModels[i].Id)
		if err != nil {
			belogs.Error("updateSyncLogFilesJsonAllAndStateDb(): updateSyncLogFileJsonAllAndState fail:",
				jsonutil.MarshalJson(syncLogFileModels[i]),
				"   syncLogFileId:", syncLogFileModels[i].Id, err)
			return err
		}
	}
	return nil
}

func updateSyncLogFileJsonAllAndStateDb(session *xorm.Session, syncLogFileModel *SyncLogFileModel) error {
	belogs.Debug("updateSyncLogFileJsonAllAndStateDb(): id:", syncLogFileModel.Id,
		"  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  syncLogFileModel:", jsonutil.MarshalJson(syncLogFileModel))
	sqlStr := `update lab_rpki_sync_log_file f set 	
	  f.state=json_replace(f.state,'$.updateCertTable','finished','$.rtr',?) ,
	  f.jsonAll=?  where f.id=?`
	rtrState := "notNeed"
	jsonAll := ""
	if (syncLogFileModel.FileType == "roa" || syncLogFileModel.FileType == "asa") &&
		syncLogFileModel.SyncType != "del" {
		rtrState = "notYet"
	}

	//when del or update(before del), syncLogFileModel.CertModel is nil
	if syncLogFileModel.CertModel == nil {
		belogs.Debug("updateSyncLogFileJsonAllAndStateDb(): del or update, CertModel is nil, syncLogFileModel:",
			jsonutil.MarshalJson(syncLogFileModel))
	} else {
		// when add or update(after del), syncLogFileModel.CertModel is not nil
		jsonAll = jsonutil.MarshalJson(syncLogFileModel.CertModel)
	}
	belogs.Debug("updateSyncLogFileJsonAllAndStateDb(): id:", syncLogFileModel.Id,
		"  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  jsonAll:", jsonAll)

	_, err := session.Exec(sqlStr, rtrState, xormdb.SqlNullString(jsonAll), syncLogFileModel.Id)
	if err != nil {
		belogs.Error("updateSyncLogFileJsonAllAndStateDb(): updateSyncLogFileJsonAllAndState fail:",
			"   id:", syncLogFileModel.Id,
			"   file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
			"   rtrState:", rtrState, "  jsonAll:", jsonAll, err)
		return err
	}
	belogs.Debug("updateSyncLogFileJsonAllAndStateDb(): update lab_rpki_sync_log_file, id:", syncLogFileModel.Id,
		"   file:", syncLogFileModel.FilePath, syncLogFileModel.FileName)
	return nil
}
