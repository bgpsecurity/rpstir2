package db

import (
	"sync"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"

	"model"
)

func UpdateSyncLogFilesJsonAllAndState(syncLogFileModels []model.SyncLogFileModel, wg *sync.WaitGroup) error {
	defer func() {
		wg.Done()
	}()
	belogs.Debug("UpdateSyncLogFilesJsonAllAndState(): len(syncLogFileModels):", len(syncLogFileModels))

	session, err := xormdb.NewSession()
	defer session.Close()
	for i, _ := range syncLogFileModels {
		rtrState := "notNeed"
		if "roa" == syncLogFileModels[i].FileType && syncLogFileModels[i].SyncType != "del" {
			rtrState = "notYet"
		}
		err := updateSyncLogFileJsonAllAndState(session, &syncLogFileModels[i], rtrState)
		if err != nil {
			belogs.Error("UpdateSyncLogFilesJsonAllAndState(): updateSyncLogFileJsonAllAndState fail:",
				jsonutil.MarshalJson(syncLogFileModels[i]),
				"   syncLogFileId:", syncLogFileModels[i].Id, err)
			return xormdb.RollbackAndLogError(session, "UpdateSyncLogFileJsonAllAndState(): updateSyncLogFileJsonAllAndState fail: "+
				jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateSyncLogFilesJsonAllAndState(): CommitSession fail :", err)
		return err
	}
	return nil

}

func updateSyncLogFileJsonAllAndState(session *xorm.Session, syncLogFileModel *model.SyncLogFileModel, rtrState string) (err error) {
	belogs.Debug("updateSyncLogFileJsonAllAndState():update lab_rpki_sync_log_file rtrState :", syncLogFileModel.Id, rtrState)

	sqlStr := `update lab_rpki_sync_log_file f set 	
	  f.state=json_replace(f.state,'$.updateCertTable','finished','$.rtr',?) ,
	  f.jsonAll=?  where f.id=?`
	_, err = session.Exec(sqlStr, rtrState, xormdb.SqlNullString(syncLogFileModel.JsonAll), syncLogFileModel.Id)
	if err != nil {
		belogs.Error("updateSyncLogFileJsonAllAndState(): update lab_rpki_sync_log_file Exec:", jsonutil.MarshalJson(syncLogFileModel),
			"   syncLogFileId:", syncLogFileModel.Id, err)
		return err
	}

	return nil
}
