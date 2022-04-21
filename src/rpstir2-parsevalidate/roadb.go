package parsevalidate

import (
	"sync"
	"time"

	model "rpstir2-model"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

// add
func addRoasDb(syncLogFileModels []SyncLogFileModel) error {
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	start := time.Now()

	belogs.Debug("addRoasDb(): len(syncLogFileModels):", len(syncLogFileModels))
	// insert new roa
	for i := range syncLogFileModels {
		err = insertRoaDb(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("addRoasDb(): insertRoaDb fail:", jsonutil.MarshalJson(syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "addRoasDb(): insertRoaDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = updateSyncLogFilesJsonAllAndStateDb(session, syncLogFileModels)
	if err != nil {
		belogs.Error("addRoasDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "addRoasDb(): updateSyncLogFilesJsonAllAndStateDb fail ", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("addRoasDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("addRoasDb(): len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

// del
func delRoasDb(delSyncLogFileModels []SyncLogFileModel, updateSyncLogFileModels []SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Info("delRoasDb(): will del len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = delRoaByIdDb(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("delRoasDb(): delRoaByIdDb fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "delRoasDb(): delRoaByIdDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = updateSyncLogFilesJsonAllAndStateDb(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("delRoasDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "delRoasDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delRoasDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delRoasDb(): len(roas), ", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func delRoaByIdDb(session *xorm.Session, roaId uint64) (err error) {

	belogs.Debug("delRoaByIdDb():delete lab_rpki_roa by roaId:", roaId)
	// rrdp may have id==0, just return nil
	if roaId <= 0 {
		return nil
	}
	belogs.Info("delRoaByIdDb():delete lab_rpki_roa by roaId, more than 0:", roaId)

	//lab_rpki_roa_ipaddress
	res, err := session.Exec("delete from lab_rpki_roa_ipaddress  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("delRoaByIdDb():delete  from lab_rpki_roa_ipaddress fail: roaId: ", roaId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("delRoaByIdDb():delete lab_rpki_roa_ipaddress by roaId:", roaId, "  count:", count)

	//lab_rpki_roa_ee_ipaddress
	res, err = session.Exec("delete from lab_rpki_roa_ee_ipaddress  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("delRoaByIdDb():delete  from lab_rpki_roa_ee_ipaddress fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delRoaByIdDb():delete lab_rpki_roa_ee_ipaddress by roaId:", roaId, "  count:", count)

	//lab_rpki_roa_sia
	res, err = session.Exec("delete from  lab_rpki_roa_sia  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("delRoaByIdDb():delete  from lab_rpki_roa_sia fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delRoaByIdDb():delete lab_rpki_roa_sia by roaId:", roaId, "  count:", count)

	//lab_rpki_roa_sia
	res, err = session.Exec("delete from  lab_rpki_roa_aia  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("delRoaByIdDb():delete  from lab_rpki_roa_aia fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delRoaByIdDb():delete lab_rpki_roa_aia by roaId:", roaId, "  count:", count)

	//lab_rpki_roa
	res, err = session.Exec("delete from  lab_rpki_roa  where id = ?", roaId)
	if err != nil {
		belogs.Error("delRoaByIdDb():delete  from lab_rpki_roa fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delRoaByIdDb():delete lab_rpki_roa by roaId:", roaId, "  count:", count)

	return nil
}

func insertRoaDb(session *xorm.Session,
	syncLogFileModel *SyncLogFileModel, now time.Time) error {

	roaModel := syncLogFileModel.CertModel.(model.RoaModel)
	//lab_rpki_roa
	sqlStr := `INSERT lab_rpki_roa(
	                asn,  ski, aki, filePath,fileName, 
	                fileHash,jsonAll,syncLogId, syncLogFileId, updateTime,
	                state)
					VALUES(?,?,?,?,?,
					?,?,?,?,?,
					?)`
	res, err := session.Exec(sqlStr,
		roaModel.Asn, xormdb.SqlNullString(roaModel.Ski), xormdb.SqlNullString(roaModel.Aki), roaModel.FilePath, roaModel.FileName,
		roaModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(roaModel)), syncLogFileModel.SyncLogId, syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)))
	if err != nil {
		belogs.Error("insertRoaDb(): INSERT lab_rpki_roa Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	roaId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("insertRoaDb(): LastInsertId :", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	//lab_rpki_roa_aia
	belogs.Debug("insertRoaDb(): roaModel.Aia.CaIssuers:", roaModel.AiaModel.CaIssuers)
	if len(roaModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT lab_rpki_roa_aia(roaId, caIssuers)
				VALUES(?,?)`
		res, err = session.Exec(sqlStr, roaId, roaModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("insertRoaDb(): INSERT lab_rpki_roa_aia Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_roa_sia
	belogs.Debug("insertRoaDb(): roaModel.Sia:", roaModel.SiaModel)
	if len(roaModel.SiaModel.CaRepository) > 0 ||
		len(roaModel.SiaModel.RpkiManifest) > 0 ||
		len(roaModel.SiaModel.RpkiNotify) > 0 ||
		len(roaModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT lab_rpki_roa_sia(roaId, rpkiManifest,rpkiNotify,caRepository,signedObject)
				VALUES(?,?,?,?,?)`
		res, err = session.Exec(sqlStr, roaId, roaModel.SiaModel.RpkiManifest,
			roaModel.SiaModel.RpkiNotify, roaModel.SiaModel.CaRepository,
			roaModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("insertRoaDb(): INSERT lab_rpki_roa_sia Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_roa_ipaddress
	belogs.Debug("insertRoaDb(): roaModel.IPAddrBlocks:", jsonutil.MarshalJson(roaModel.RoaIpAddressModels))
	if roaModel.RoaIpAddressModels != nil && len(roaModel.RoaIpAddressModels) > 0 {
		sqlStr = `INSERT lab_rpki_roa_ipaddress(roaId, addressFamily,addressPrefix,maxLength, rangeStart, rangeEnd,addressPrefixRange )
						VALUES(?,?,?,?,?,?,?)`
		for _, roaIpAddressModel := range roaModel.RoaIpAddressModels {
			res, err = session.Exec(sqlStr, roaId, roaIpAddressModel.AddressFamily,
				roaIpAddressModel.AddressPrefix, roaIpAddressModel.MaxLength,
				roaIpAddressModel.RangeStart, roaIpAddressModel.RangeEnd, roaIpAddressModel.AddressPrefixRange)
			if err != nil {
				belogs.Error("insertRoaDb(): INSERT lab_rpki_roa_ipaddress Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
				return err
			}

		}
	}

	//lab_rpki_roa_ee_ipaddress
	belogs.Debug("insertRoaDb(): roaModel.CerIpAddressModel:", roaModel.EeCertModel.CerIpAddressModel)
	sqlStr = `INSERT lab_rpki_roa_ee_ipaddress(roaId,addressFamily, addressPrefix,min,max,
	                rangeStart,rangeEnd,addressPrefixRange) 
	                 VALUES(?,?,?,?,?,
	                 ?,?,?)`
	for _, cerIpAddress := range roaModel.EeCertModel.CerIpAddressModel.CerIpAddresses {
		res, err = session.Exec(sqlStr,
			roaId, cerIpAddress.AddressFamily, cerIpAddress.AddressPrefix, cerIpAddress.Min, cerIpAddress.Max,
			cerIpAddress.RangeStart, cerIpAddress.RangeEnd, cerIpAddress.AddressPrefixRange)
		if err != nil {
			belogs.Error("insertRoaDb(): INSERT lab_rpki_roa_ee_ipaddress Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}
	return nil
}

func getExpireRoaDb(now time.Time) (certIdStateModels []CertIdStateModel, err error) {

	certIdStateModels = make([]CertIdStateModel, 0)
	t := now.Local().Format("2006-01-02T15:04:05-0700")
	sql := `select id, state as stateStr,str_to_date( SUBSTRING_INDEX(c.jsonAll->>'$.eeCertModel.notAfter','+',1),'%Y-%m-%dT%H:%i:%S')  as endTime  from  lab_rpki_roa c 
			where c.jsonAll->>'$.eeCertModel.notAfter' < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireRoaDb(): lab_rpki_roa fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireRoaDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateRoaStateDb(certIdStateModels []CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	sql := `update lab_rpki_roa c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateRoaStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateRoaStateDb(): UPDATE lab_rpki_roa fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateRoaStateDb(): UPDATE lab_rpki_roa fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateRoaStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateRoaStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Now().Sub(start))

	return nil
}
