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
func AddMfts(syncLogFileModels []parsevalidatemodel.SyncLogFileModel) error {
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	start := time.Now()

	belogs.Debug("AddMfts(): len(syncLogFileModels):", len(syncLogFileModels))
	// insert new mft
	for i := range syncLogFileModels {
		err = insertMft(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("AddMfts(): insertMft fail:", jsonutil.MarshalJson(syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "AddMfts(): insertMft fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = UpdateSyncLogFilesJsonAllAndState(session, syncLogFileModels)
	if err != nil {
		belogs.Error("AddMfts(): UpdateSyncLogFilesJsonAllAndState fail:", err)
		return xormdb.RollbackAndLogError(session, "AddMfts(): UpdateSyncLogFilesJsonAllAndState fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("AddMfts(): insertMft CommitSession fail :", err)
		return err
	}
	belogs.Info("AddMfts(): len(mfts):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

// del
func DelMfts(delSyncLogFileModels []parsevalidatemodel.SyncLogFileModel, updateSyncLogFileModels []parsevalidatemodel.SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()

	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Debug("DelMfts(): len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = delMftById(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("DelMfts(): DelMftByFile fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "DelMfts(): DelMftById fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = UpdateSyncLogFilesJsonAllAndState(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("DelMfts(): UpdateSyncLogFilesJsonAllAndState fail:", err)
		return xormdb.RollbackAndLogError(session, "DelMfts(): UpdateSyncLogFilesJsonAllAndState fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelMfts(): CommitSession fail :", err)
		return err
	}
	belogs.Info("DelMfts(): len(mfts):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
func DelMftByFile(session *xorm.Session, filePath, fileName string) (err error) {
	// try to delete old
	belogs.Debug("DelMftByFile():will delete lab_rpki_mft by filePath+fileName:", filePath, fileName)

	labRpkiMft := model.LabRpkiMft{}
	var mftId uint64
	has, err := session.Table(&labRpkiMft).Where("filePath=?", filePath).And("fileName=?", fileName).Cols("id").Get(&mftId)
	if err != nil {
		belogs.Error("DelMftByFile(): get current labRpkiMft fail:", filePath, fileName, err)
		return err
	}

	belogs.Debug("DelMftByFile():will delete lab_rpki_mft mftId:", mftId, "    has:", has)
	if has {
		return delMftById(session, mftId)
	}
	return nil
}

func delMftById(session *xorm.Session, mftId uint64) (err error) {
	belogs.Debug("delMftById():delete lab_rpki_mft_ by mftId:", mftId)

	// rrdp may have id==0, just return nil
	if mftId <= 0 {
		return nil
	}

	//lab_rpki_mft_file_hash
	_, err = session.Exec("delete from lab_rpki_mft_file_hash  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("delMftById():delete  from lab_rpki_mft_file_hash fail: mftId: ", mftId, err)
		return err
	}

	//lab_rpki_mft_sia
	_, err = session.Exec("delete from  lab_rpki_mft_sia  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("delMftById():delete  from lab_rpki_mft_sia fail:mftId: ", mftId, err)
		return err
	}

	//lab_rpki_mft_aia
	_, err = session.Exec("delete from  lab_rpki_mft_aia  where mftId = ?", mftId)
	if err != nil {
		belogs.Error("delMftById():delete  from lab_rpki_mft_aia fail:mftId: ", mftId, err)
		return err
	}

	//lab_rpki_mft
	_, err = session.Exec("delete from  lab_rpki_mft  where id = ?", mftId)
	if err != nil {
		belogs.Error("delMftById():delete  from lab_rpki_mft fail:mftId: ", mftId, err)
		return err
	}

	return nil

}

func insertMft(session *xorm.Session,
	syncLogFileModel *parsevalidatemodel.SyncLogFileModel, now time.Time) error {

	mftModel := syncLogFileModel.CertModel.(model.MftModel)
	thisUpdate := mftModel.ThisUpdate
	nextUpdate := mftModel.NextUpdate
	belogs.Debug("insertMft():now ", now, "  thisUpdate:", thisUpdate, "  nextUpdate:", nextUpdate, "    mftModel:", jsonutil.MarshalJson(mftModel))

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
		belogs.Error("insertMft(): INSERT lab_rpki_mft Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	mftId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("insertMft(): LastInsertId :", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	//lab_rpki_mft_aia
	belogs.Debug("insertMft(): mftModel.Aia.CaIssuers:", mftModel.AiaModel.CaIssuers)
	if len(mftModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT lab_rpki_mft_aia(mftId, caIssuers) 
			VALUES(?,?)`
		res, err = session.Exec(sqlStr, mftId, mftModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("insertMft(): INSERT lab_rpki_mft_aia Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_mft_sia
	belogs.Debug("insertMft(): mftModel.Sia:", mftModel.SiaModel)
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
			belogs.Error("insertMft(): INSERT lab_rpki_mft_sia Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_mft_fileAndHashs
	belogs.Debug("insertMft(): mftModel.FileHashModels:", mftModel.FileHashModels)
	if mftModel.FileHashModels != nil && len(mftModel.FileHashModels) > 0 {
		sqlStr = `INSERT lab_rpki_mft_file_hash(mftId, file,hash) VALUES(?,?,?)`
		for _, fileHashModel := range mftModel.FileHashModels {
			res, err = session.Exec(sqlStr, mftId, fileHashModel.File, fileHashModel.Hash)
			if err != nil {
				belogs.Error("insertMft(): INSERT lab_rpki_mft_file_hash Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
				return err
			}
		}
	}
	return nil
}
