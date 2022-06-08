package parsevalidate

import (
	"sync"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
	"xorm.io/xorm"
)

// add
func addCrlsDb(syncLogFileModels []SyncLogFileModel) error {
	belogs.Info("addCrlsDb(): will insert len(syncLogFileModels):", len(syncLogFileModels))
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	start := time.Now()

	// add
	belogs.Debug("addCrlsDb(): len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = insertCrlDb(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("addCrlsDb(): insertCrlDb fail:", jsonutil.MarshalJson(syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "addCrlsDb(): insertCrlDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = updateSyncLogFilesJsonAllAndStateDb(session, syncLogFileModels)
	if err != nil {
		belogs.Error("addCrlsDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "addCrlsDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("addCrlsDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("addCrlsDb(): len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil

}

// del
func delCrlsDb(delSyncLogFileModels []SyncLogFileModel, updateSyncLogFileModels []SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()

	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Info("delCrlsDb(): will del len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = delCrlByIdDb(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("delCrlsDb(): delCrlByIdDb fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "delCrlsDb(): delCrlByIdDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = updateSyncLogFilesJsonAllAndStateDb(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("delCrlsDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "delCrlsDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delCrlsDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delCrlsDb(): len(crls):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func delCrlByIdDb(session *xorm.Session, crlId uint64) (err error) {
	belogs.Debug("delCrlByIdDb():delete lab_rpki_crl by crlId:", crlId)

	// rrdp may have id==0, just return nil
	if crlId <= 0 {
		return nil
	}
	belogs.Info("delCrlByIdDb():delete lab_rpki_crl by crlId, more than 0:", crlId)

	//lab_rpki_crl_revoked_cert
	res, err := session.Exec("delete from lab_rpki_crl_revoked_cert  where crlId = ?", crlId)
	if err != nil {
		belogs.Error("delCrlByIdDb():delete  from lab_rpki_crl_revoked_cert fail: crlId: ", crlId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("delCrlByIdDb():delete lab_rpki_crl_revoked_cert by crlId:", crlId, "  count:", count)

	//lab_rpki_crl_revoked
	res, err = session.Exec("delete from  lab_rpki_crl  where id = ?", crlId)
	if err != nil {
		belogs.Error("delCrlByIdDb():delete  from lab_rpki_crl fail: crlId: ", crlId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delCrlByIdDb():delete lab_rpki_crl by crlId:", crlId, "  count:", count)

	return nil

}

func insertCrlDb(session *xorm.Session,
	syncLogFileModel *SyncLogFileModel, now time.Time) error {

	crlModel := syncLogFileModel.CertModel.(model.CrlModel)
	thisUpdate := crlModel.ThisUpdate
	nextUpdate := crlModel.NextUpdate
	belogs.Debug("insertCrlDb(): crlModel:", jsonutil.MarshalJson(crlModel), "  now ", now)

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
		belogs.Error("insertCrlDb(): INSERT lab_rpki_crl Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	crlId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("insertCrlDb(): LastInsertId :", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	//lab_rpki_crl_crlrevokedcerts
	belogs.Debug("insertCrlDb(): crlModel.RevokedCertModels:", crlModel.RevokedCertModels)
	if crlModel.RevokedCertModels != nil && len(crlModel.RevokedCertModels) > 0 {
		sqlStr = `INSERT lab_rpki_crl_revoked_cert(crlId, sn, revocationTime) VALUES(?,?,?)`
		for _, revokedCertModel := range crlModel.RevokedCertModels {
			res, err = session.Exec(sqlStr, crlId, revokedCertModel.Sn, revokedCertModel.RevocationTime)
			if err != nil {
				belogs.Error("insertCrlDb(): INSERT lab_rpki_crl_revoked_cert Exec :",
					jsonutil.MarshalJson(syncLogFileModel), err)
				return err
			}
		}
	}
	return nil
}

func getExpireCrlDb(now time.Time) (certIdStateModels []CertIdStateModel, err error) {

	certIdStateModels = make([]CertIdStateModel, 0)
	t := convert.Time2String(now)
	sql := `select id, state as stateStr, c.nextUpdate  as endTime  from  lab_rpki_crl c 
			where timestamp(c.nextUpdate) < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireCrlDb(): lab_rpki_crl fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireCrlDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateCrlStateDb(certIdStateModels []CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	sql := `update lab_rpki_crl c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateCrlStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateCrlStateDb(): UPDATE lab_rpki_crl fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateCrlStateDb(): UPDATE lab_rpki_crl fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateCrlStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateCrlStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Now().Sub(start))

	return nil
}
