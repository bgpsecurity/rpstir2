package db

import (
	"sync"
	"time"

	belogs "github.com/astaxie/beego/logs"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	xormdb "github.com/cpusoft/goutil/xormdb"
	"github.com/go-xorm/xorm"

	"model"
	parsevalidatemodel "parsevalidate/model"
)

// add
func AddCrls(syncLogFileModels []parsevalidatemodel.SyncLogFileModel) error {
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	start := time.Now()

	// add
	belogs.Debug("AddCrls(): len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = insertCrl(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("AddCrls(): insertCrl fail:", jsonutil.MarshalJson(syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "AddCrls(): insertCrl fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = UpdateSyncLogFilesJsonAllAndState(session, syncLogFileModels)
	if err != nil {
		belogs.Error("AddCrls(): UpdateSyncLogFilesJsonAllAndState fail:", err)
		return xormdb.RollbackAndLogError(session, "AddCrls(): UpdateSyncLogFilesJsonAllAndState fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("AddCrls(): insertCrl CommitSession fail :", err)
		return err
	}
	belogs.Info("AddCrls(): len(crls):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil

}

// del
func DelCrls(delSyncLogFileModels []parsevalidatemodel.SyncLogFileModel, updateSyncLogFileModels []parsevalidatemodel.SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()

	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Debug("DelCrls(): len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = delCrlById(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("DelCrls(): DelCrlByFile fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "DelCrls(): DelCrlById fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = UpdateSyncLogFilesJsonAllAndState(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("DelCrls(): UpdateSyncLogFilesJsonAllAndState fail:", err)
		return xormdb.RollbackAndLogError(session, "DelCrls(): UpdateSyncLogFilesJsonAllAndState fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelCrls(): CommitSession fail :", err)
		return err
	}
	belogs.Info("DelCrls(): len(crls):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
func DelCrlByFile(session *xorm.Session, filePath, fileName string) (err error) {
	// try to delete old
	belogs.Debug("DelCrlByFile():will delete lab_rpki_crl by filePath+fileName:", filePath, fileName)
	labRpkiCrl := model.LabRpkiCrl{}
	var crlId uint64
	has, err := session.Table(&labRpkiCrl).Where("filePath=?", filePath).And("fileName=?", fileName).Cols("id").Get(&crlId)
	if err != nil {
		belogs.Error("DelCrlByFile(): get current labRpkiCrl fail:", filePath, fileName, err)
		return err
	}

	belogs.Debug("DelCrlByFile():will delete lab_rpki_crl by crlId:", crlId, "    has:", has)
	if has {
		return delCrlById(session, crlId)
	}
	return nil
}
func delCrlById(session *xorm.Session, crlId uint64) (err error) {
	belogs.Debug("delCrlById(): crlId:", crlId)

	// rrdp may have id==0, just return nil
	if crlId <= 0 {
		return nil
	}

	//lab_rpki_crl_revoked_cert
	_, err = session.Exec("delete from lab_rpki_crl_revoked_cert  where crlId = ?", crlId)
	if err != nil {
		belogs.Error("delCrlById():delete  from lab_rpki_crl_revoked_cert fail: crlId: ", crlId, err)
		return err
	}

	//lab_rpki_crl_revoked
	_, err = session.Exec("delete from  lab_rpki_crl  where id = ?", crlId)
	if err != nil {
		belogs.Error("delCrlById():delete  from lab_rpki_crl fail: crlId: ", crlId, err)
		return err
	}
	return nil

}

func insertCrl(session *xorm.Session,
	syncLogFileModel *parsevalidatemodel.SyncLogFileModel, now time.Time) error {

	crlModel := syncLogFileModel.CertModel.(model.CrlModel)
	thisUpdate := crlModel.ThisUpdate
	nextUpdate := crlModel.NextUpdate
	belogs.Debug("insertCrl(): crlModel:", jsonutil.MarshalJson(crlModel), "  now ", now)

	//lab_rpki_crl
	sqlStr := `INSERT lab_rpki_crl(
	        crlNumber, thisUpdate, nextUpdate, hasExpired, aki, 
	        filePath,fileName,fileHash, jsonAll,syncLogId, 
	        syncLogFileId, updateTime, state)
			VALUES(?,?,?,?,?,
			?,?,?,?,?,
			?,?,?)`
	res, err := session.Exec(sqlStr,
		crlModel.CrlNumber, thisUpdate, nextUpdate, crlModel.HasExpired, xormdb.SqlNullString(crlModel.Aki),
		crlModel.FilePath, crlModel.FileName, crlModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(crlModel)), syncLogFileModel.SyncLogId,
		syncLogFileModel.Id, now, xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)))
	if err != nil {
		belogs.Error("insertCrl(): INSERT lab_rpki_crl Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	crlId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("insertCrl(): LastInsertId :", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	//lab_rpki_crl_crlrevokedcerts
	belogs.Debug("insertCrl(): crlModel.RevokedCertModels:", crlModel.RevokedCertModels)
	if crlModel.RevokedCertModels != nil && len(crlModel.RevokedCertModels) > 0 {
		sqlStr = `INSERT lab_rpki_crl_revoked_cert(crlId, sn, revocationTime) VALUES(?,?,?)`
		for _, revokedCertModel := range crlModel.RevokedCertModels {
			res, err = session.Exec(sqlStr, crlId, revokedCertModel.Sn, revokedCertModel.RevocationTime)
			if err != nil {
				belogs.Error("insertCrl(): INSERT lab_rpki_crl_revoked_cert Exec :",
					jsonutil.MarshalJson(syncLogFileModel), err)
				return err
			}
		}
	}
	return nil
}
