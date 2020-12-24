package db

import (
	"time"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"

	parsevalidatemodel "parsevalidate/model"
)

func GetSyncLogFileModelsBySyncLogId(labRpkiSyncLogId uint64) (syncLogFileModels *parsevalidatemodel.SyncLogFileModels, err error) {
	start := time.Now()

	belogs.Debug("GetSyncLogFileModelsBySyncLogId():start")
	dbSyncLogFileModels := make([]parsevalidatemodel.SyncLogFileModel, 0)
	sql := `select s.id,s.syncLogId,s.filePath,s.fileName, s.fileType, s.syncType, 
				cast(CONCAT(IFNULL(c.id,''),IFNULL(m.id,''),IFNULL(l.id,''),IFNULL(r.id,'')) as unsigned int) as certId from lab_rpki_sync_log_file s 
			left join lab_rpki_cer c on c.filePath = s.filePath and c.fileName = s.fileName  
			left join lab_rpki_mft m on m.filePath = s.filePath and m.fileName = s.fileName  
			left join lab_rpki_crl l on l.filePath = s.filePath and l.fileName = s.fileName  
			left join lab_rpki_roa r on r.filePath = s.filePath and r.fileName = s.fileName 
			where s.state->>'$.updateCertTable'='notYet' and s.syncLogId=? order by s.id `
	err = xormdb.XormEngine.SQL(sql, labRpkiSyncLogId).Find(&dbSyncLogFileModels)
	if err != nil {
		belogs.Error("GetSyncLogFileModelsBySyncLogId(): Find fail:", err)
		return nil, err
	}
	belogs.Debug("GetSyncLogFileModelsBySyncLogId(): len(dbSyncLogFileModels):", len(dbSyncLogFileModels), jsonutil.MarshalJson(dbSyncLogFileModels))
	syncLogFileModels = parsevalidatemodel.NewSyncLogFileModels(labRpkiSyncLogId, dbSyncLogFileModels)
	belogs.Info("GetSyncLogFileModelsBySyncLogId(): end, len(dbSyncLogFileModels),  time(s):", len(dbSyncLogFileModels), time.Now().Sub(start).Seconds())
	return syncLogFileModels, nil

}

func UpdateSyncLogFilesJsonAllAndState(session *xorm.Session, syncLogFileModels []parsevalidatemodel.SyncLogFileModel) error {
	belogs.Debug("UpdateSyncLogFilesJsonAllAndState(): len(syncLogFileModels):", len(syncLogFileModels))
	sqlStr := `update lab_rpki_sync_log_file f set 	
	  f.state=json_replace(f.state,'$.updateCertTable','finished','$.rtr',?) ,
	  f.jsonAll=?  where f.id=?`
	for i := range syncLogFileModels {
		rtrState := "notNeed"
		jsonAll := ""
		if syncLogFileModels[i].FileType == "roa" && syncLogFileModels[i].SyncType != "del" {
			rtrState = "notYet"
		}

		//when del or update(before del), syncLogFileModels[i].CertModel is nil
		if syncLogFileModels[i].CertModel == nil {
			belogs.Debug("UpdateSyncLogFilesJsonAllAndState(): del or update, CertModel is nil:",
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
					belogs.Error("UpdateSyncLogFilesJsonAllAndState(): syncLogFileModels[i].FileType fail:",
						syncLogFileModels[i].FileType)
					return errors.New("syncLogFileModels[i].FileType fail, " + syncLogFileModels[i].FileType)
				}
			*/
			jsonAll = jsonutil.MarshalJson(syncLogFileModels[i].CertModel)
		}

		_, err := session.Exec(sqlStr, rtrState, xormdb.SqlNullString(jsonAll), syncLogFileModels[i].Id)
		if err != nil {
			belogs.Error("UpdateSyncLogFilesJsonAllAndState(): updateSyncLogFileJsonAllAndState fail:",
				jsonutil.MarshalJson(syncLogFileModels[i]),
				"   syncLogFileId:", syncLogFileModels[i].Id, err)
			return err
		}
	}
	return nil
}
