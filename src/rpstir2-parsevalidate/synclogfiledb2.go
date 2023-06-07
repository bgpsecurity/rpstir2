package parsevalidate

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

// need endCh, when error
func getSyncLogFileModelBySyncLogIdDb(labRpkiSyncLogId uint64, syncLogFileModelCh chan *SyncLogFileModel,
	parseConcurrentCh chan int, endCh chan bool, wg *sync.WaitGroup) (err error) {
	start := time.Now()

	belogs.Debug("getSyncLogFileModelBySyncLogIdDb(): labRpkiSyncLogId:", labRpkiSyncLogId)
	syncLogFileModel := new(SyncLogFileModel)
	sql := `select s.id,s.syncLogId,s.filePath,s.fileName, s.fileType, s.syncType, 
				cast(CONCAT(IFNULL(c.id,''),IFNULL(m.id,''),IFNULL(l.id,''),IFNULL(r.id,''),IFNULL(a.id,'')) as unsigned int) as certId from lab_rpki_sync_log_file s 
			left join lab_rpki_cer c on c.filePath = s.filePath and c.fileName = s.fileName  
			left join lab_rpki_mft m on m.filePath = s.filePath and m.fileName = s.fileName  
			left join lab_rpki_crl l on l.filePath = s.filePath and l.fileName = s.fileName  
			left join lab_rpki_roa r on r.filePath = s.filePath and r.fileName = s.fileName 
			left join lab_rpki_asa a on a.filePath = s.filePath and a.fileName = s.fileName 
			where s.state->>'$.updateCertTable'='notYet' and s.syncLogId=? order by s.id `
	rows, err := xormdb.XormEngine.SQL(sql, labRpkiSyncLogId).Rows(syncLogFileModel)
	if err != nil {
		belogs.Error("getSyncLogFileModelBySyncLogIdDb(): select from rpki_*** fail:", err)
		return err
	}
	belogs.Debug("getSyncLogFileModelBySyncLogIdDb(): will call rows.Next(), time(s):", time.Since(start))

	defer rows.Close()
	var index uint64
	for rows.Next() {
		// control parse speed
		parseConcurrentCh <- 1
		// get new *syncLogFileModel every Scan
		syncLogFileModel = new(SyncLogFileModel)
		err = rows.Scan(syncLogFileModel)
		if err != nil {
			belogs.Error("getSyncLogFileModelBySyncLogIdDb(): Scan fail:", err)
			continue
		}
		syncLogFileModel.Index = index
		belogs.Debug("getSyncLogFileModelBySyncLogIdDb(): Scan, wg.Add() id:", syncLogFileModel.Id, " index:", index,
			"  file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
			"  , time(s):", time.Since(start))

		atomic.AddInt64(&parseCount, 1)
		wg.Add(1)
		belogs.Debug("getSyncLogFileModelBySyncLogIdDb(): AddInt64(1), parseCount:", atomic.LoadInt64(&parseCount))

		syncLogFileModelCh <- syncLogFileModel
		index++
	}

	belogs.Info("getSyncLogFileModelBySyncLogIdDb(): get all syncLogFileModel,labRpkiSyncLogId:", labRpkiSyncLogId, "   count:", index, "  time(s):", time.Since(start))
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
