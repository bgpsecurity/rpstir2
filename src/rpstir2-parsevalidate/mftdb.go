package parsevalidate

import (
	"sync"
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/convert"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

// add
func addMftsDb(syncLogFileModels []SyncLogFileModel) error {
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	start := time.Now()

	belogs.Debug("addMftsDb(): len(syncLogFileModels):", len(syncLogFileModels))
	// insert new mft
	for i := range syncLogFileModels {
		err = insertMftDb(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("addMftsDb(): insertMftDb fail:", jsonutil.MarshalJson(syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "addMftsDb(): insertMftDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = updateSyncLogFilesJsonAllAndStateDb(session, syncLogFileModels)
	if err != nil {
		belogs.Error("addMftsDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "addMftsDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("addMftsDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("addMftsDb(): len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

// del
func delMftsDb(delSyncLogFileModels []SyncLogFileModel, updateSyncLogFileModels []SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()

	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Debug("delMftsDb(): len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = delMftByIdDb(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("delMftsDb(): delMftByIdDb fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "delMftsDb(): delMftByIdDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = updateSyncLogFilesJsonAllAndStateDb(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("delMftsDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "delMftsDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delMftsDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delMftsDb(): len(mfts):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func delMftByIdDb(session *xorm.Session, mftId uint64) (err error) {
	belogs.Info("delMftByIdDb():delete lab_rpki_mft by mftId:", mftId)

	// rrdp may have id==0, just return nil
	if mftId <= 0 {
		return nil
	}

	//lab_rpki_mft_file_hash
	res, err := session.Exec("delete from lab_rpki_mft_file_hash  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("delMftByIdDb():delete  from lab_rpki_mft_file_hash fail: mftId: ", mftId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("delMftByIdDb():delete lab_rpki_mft_file_hash by mftId:", mftId, "  count:", count)

	//lab_rpki_mft_sia
	res, err = session.Exec("delete from  lab_rpki_mft_sia  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("delMftByIdDb():delete  from lab_rpki_mft_sia fail:mftId: ", mftId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delMftByIdDb():delete lab_rpki_mft_sia by mftId:", mftId, "  count:", count)

	//lab_rpki_mft_aia
	res, err = session.Exec("delete from  lab_rpki_mft_aia  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("delMftByIdDb():delete  from lab_rpki_mft_aia fail:mftId: ", mftId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delMftByIdDb():delete lab_rpki_mft_aia by mftId:", mftId, "  count:", count)

	//lab_rpki_mft
	res, err = session.Exec("delete from  lab_rpki_mft  where id = ?", mftId)
	if err != nil {
		belogs.Error("delMftByIdDb():delete  from lab_rpki_mft fail:mftId: ", mftId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delMftByIdDb():delete lab_rpki_mft by mftId:", mftId, "  count:", count)

	return nil

}

func insertMftDb(session *xorm.Session,
	syncLogFileModel *SyncLogFileModel, now time.Time) error {

	mftModel := syncLogFileModel.CertModel.(model.MftModel)
	thisUpdate := mftModel.ThisUpdate
	nextUpdate := mftModel.NextUpdate
	belogs.Debug("insertMftDb():now ", now, "  thisUpdate:", thisUpdate, "  nextUpdate:", nextUpdate, "    mftModel:", jsonutil.MarshalJson(mftModel))

	//lab_rpki_manifest
	sqlStr := `INSERT lab_rpki_mft(
	           mftNumber, thisUpdate, nextUpdate, ski, aki, 
	           filePath,fileName,fileHash, jsonAll,syncLogId, 
	           syncLogFileId, updateTime,state)
				VALUES(
				?,?,?,?,?,
				?,?,?,?,?,
				?,?,?)`
	res, err := session.Exec(sqlStr,
		mftModel.MftNumber, thisUpdate, nextUpdate, xormdb.SqlNullString(mftModel.Ski), xormdb.SqlNullString(mftModel.Aki),
		mftModel.FilePath, mftModel.FileName, mftModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(mftModel)), syncLogFileModel.SyncLogId,
		syncLogFileModel.Id, now, xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)))
	if err != nil {
		belogs.Error("insertMftDb(): INSERT lab_rpki_mft Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	mftId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("insertMftDb(): LastInsertId :", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	//lab_rpki_mft_aia
	belogs.Debug("insertMftDb(): mftModel.Aia.CaIssuers:", mftModel.AiaModel.CaIssuers)
	if len(mftModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT lab_rpki_mft_aia(mftId, caIssuers) 
			VALUES(?,?)`
		res, err = session.Exec(sqlStr, mftId, mftModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("insertMftDb(): INSERT lab_rpki_mft_aia Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_mft_sia
	belogs.Debug("insertMftDb(): mftModel.Sia:", mftModel.SiaModel)
	if len(mftModel.SiaModel.CaRepository) > 0 ||
		len(mftModel.SiaModel.RpkiManifest) > 0 ||
		len(mftModel.SiaModel.RpkiNotify) > 0 ||
		len(mftModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT lab_rpki_mft_sia(mftId, rpkiManifest,rpkiNotify,caRepository,signedObject) 
			VALUES(?,?,?,?,?)`
		res, err = session.Exec(sqlStr, mftId, mftModel.SiaModel.RpkiManifest,
			mftModel.SiaModel.RpkiNotify, mftModel.SiaModel.CaRepository,
			mftModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("insertMftDb(): INSERT lab_rpki_mft_sia Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_mft_fileAndHashs
	belogs.Debug("insertMftDb(): mftModel.FileHashModels:", mftModel.FileHashModels)
	if mftModel.FileHashModels != nil && len(mftModel.FileHashModels) > 0 {
		sqlStr = `INSERT lab_rpki_mft_file_hash(mftId, file,hash) VALUES(?,?,?)`
		for _, fileHashModel := range mftModel.FileHashModels {
			res, err = session.Exec(sqlStr, mftId, fileHashModel.File, fileHashModel.Hash)
			if err != nil {
				belogs.Error("insertMftDb(): INSERT lab_rpki_mft_file_hash Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
				return err
			}
		}
	}
	return nil
}

func getExpireMftDb(now time.Time) (certIdStateModels []CertIdStateModel, err error) {

	certIdStateModels = make([]CertIdStateModel, 0)
	t := convert.Time2String(now)
	sql := `select id, state as stateStr, c.nextUpdate  as endTime from  lab_rpki_mft c 
			where timestamp(c.nextUpdate) < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireMftDb(): lab_rpki_mft fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireMftDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateMftStateDb(certIdStateModels []CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	sql := `update lab_rpki_mft c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateMftStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateMftStateDb(): UPDATE lab_rpki_mft fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateMftStateDb(): UPDATE lab_rpki_mft fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateMftStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateMftStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Now().Sub(start))

	return nil
}
