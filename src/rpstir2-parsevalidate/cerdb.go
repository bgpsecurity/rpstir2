package parsevalidate

import (
	"errors"
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
func addCersDb(syncLogFileModels []SyncLogFileModel) error {
	belogs.Info("addCersDb(): will insert len(syncLogFileModels):", len(syncLogFileModels))
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	start := time.Now()

	belogs.Debug("addCersDb(): len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		// insert new cer
		err = insertCerDb(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("addCersDb(): insertCerDb fail:", jsonutil.MarshalJson(&syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "addCersDb(): insertCerDb fail: "+jsonutil.MarshalJson(&syncLogFileModels[i]), err)
		}
	}

	err = updateSyncLogFilesJsonAllAndStateDb(session, syncLogFileModels)
	if err != nil {
		belogs.Error("addCersDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "addCersDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("addCersDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("addCersDb(): len(syncLogFileModels):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil

}

// del
func delCersDb(delSyncLogFileModels []SyncLogFileModel, updateSyncLogFileModels []SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()

	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Info("delCersDb(): will del len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = delCerByIdDb(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("delCersDb(): delCerByIdDb fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "delCersDb(): delCerByIdDb fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = updateSyncLogFilesJsonAllAndStateDb(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("delCersDb(): updateSyncLogFilesJsonAllAndStateDb fail:", err)
		return xormdb.RollbackAndLogError(session, "delCersDb(): updateSyncLogFilesJsonAllAndStateDb fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("delCersDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("delCersDb(): len(cers):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func delCerByIdDb(session *xorm.Session, cerId uint64) (err error) {

	belogs.Debug("delCerByIdDb():delete lab_rpki_cer by cerId:", cerId)
	// rrdp may have id==0, just return nil
	if cerId <= 0 {
		return nil
	}
	belogs.Info("delCerByIdDb():delete lab_rpki_cer by cerId, more than 0:", cerId)

	//lab_rpki_cer_sia
	res, err := session.Exec("delete from lab_rpki_cer_sia  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("delCerByIdDb():delete  from lab_rpki_cer_sia failed, cerId:", cerId, "    err:", err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("delCerByIdDb():delete lab_rpki_cer_sia by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_ipaddress
	res, err = session.Exec("delete from  lab_rpki_cer_ipaddress  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("delCerByIdDb():delete  from lab_rpki_cer_ipaddress failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delCerByIdDb():delete lab_rpki_cer_ipaddress by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_crldp
	res, err = session.Exec("delete  from lab_rpki_cer_crldp  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("delCerByIdDb():delete  from lab_rpki_cer_crldp failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delCerByIdDb():delete lab_rpki_cer_crldp by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_asn
	res, err = session.Exec("delete  from lab_rpki_cer_asn  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("delCerByIdDb():delete  from lab_rpki_cer_asn  failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delCerByIdDb():delete lab_rpki_cer_asn by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_aia
	res, err = session.Exec("delete  from lab_rpki_cer_aia  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("delCerByIdDb():delete  from lab_rpki_cer_aia  failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delCerByIdDb():delete lab_rpki_cer_aia by cerId:", cerId, "  count:", count)

	//lab_rpki_cer
	res, err = session.Exec("delete  from lab_rpki_cer  where id = ?", cerId)
	if err != nil {
		belogs.Error("delCerByIdDb():delete  from lab_rpki_cer  failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("delCerByIdDb():delete lab_rpki_cer by cerId:", cerId, "  count:", count)

	return nil
}

func insertCerDb(session *xorm.Session,
	syncLogFileModel *SyncLogFileModel, now time.Time) error {

	cerModel := syncLogFileModel.CertModel.(model.CerModel)
	notBefore := cerModel.NotBefore
	notAfter := cerModel.NotAfter
	belogs.Debug("insertCerDb():now ", now, "  notBefore:", notBefore, "  notAfter:", notAfter,
		"    cerModel:", jsonutil.MarshalJson(cerModel))

	//lab_rpki_cer
	sqlStr := `INSERT lab_rpki_cer(
	    sn, notBefore,notAfter,subject,
	    issuer,ski,aki,filePath,fileName,
	    fileHash,jsonAll,syncLogId,syncLogFileId,updateTime,
	    state) 	
	    VALUES(?,?,?,?,
	    ?,?,?,?,?,
	    ?,?,?,?,?,
	    ?)`
	res, err := session.Exec(sqlStr,
		cerModel.Sn, notBefore, notAfter, cerModel.Subject,
		cerModel.Issuer, xormdb.SqlNullString(cerModel.Ski), xormdb.SqlNullString(cerModel.Aki), cerModel.FilePath, cerModel.FileName,
		cerModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(cerModel)), syncLogFileModel.SyncLogId, syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)))
	if err != nil {
		belogs.Error("insertCerDb(): INSERT lab_rpki_cer Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	cerId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("insertCerDb(): LastInsertId:", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}
	belogs.Debug("insertCerDb():LastInsertId cerId:", cerId)

	//lab_rpki_cer_aia
	belogs.Debug("insertCerDb(): cerModel.Aia.CaIssuers:", cerModel.AiaModel.CaIssuers)
	if len(cerModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT lab_rpki_cer_aia(cerId, caIssuers) VALUES(?,?)`
		res, err = session.Exec(sqlStr, cerId, cerModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("insertCerDb(): INSERT lab_rpki_cer_aia Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_cer_asn
	belogs.Debug("insertCerDb(): cerModel.Asn:", cerModel.AsnModel)
	if len(cerModel.AsnModel.Asns) > 0 {
		sqlAsnStr := `INSERT lab_rpki_cer_asn(cerId, asn) VALUES(?,?)`
		sqlMinMaxStr := `INSERT lab_rpki_cer_asn(cerId, min,max) VALUES(?,?,?)`
		for _, asn := range cerModel.AsnModel.Asns {
			// need  asNum >=0
			if asn.Asn >= 0 {
				res, err = session.Exec(sqlAsnStr, cerId, asn.Asn)
				if err != nil {
					belogs.Error("insertCerDb(): INSERT sqlAsnStr lab_rpki_cer_asn ,syncLogFileModel err:", jsonutil.MarshalJson(syncLogFileModel), err)
					return err
				}
			} else if asn.Max >= 0 && asn.Min >= 0 {
				res, err = session.Exec(sqlMinMaxStr, cerId, asn.Min, asn.Max)
				if err != nil {
					belogs.Error("insertCerDb(): INSERT sqlMinMaxStr lab_rpki_cer_asn,syncLogFileModel err:", jsonutil.MarshalJson(syncLogFileModel), err)
					return err
				}
			} else {
				belogs.Error("insertCerDb(): INSERT lab_rpki_cer_asn asn/min/max all are zero, syncLogFileModel err:", jsonutil.MarshalJson(syncLogFileModel))
				return errors.New("insert lab_rpki_cer_asn fail, asn/min/max all are zero")
			}
		}
	}

	//lab_rpki_cer_crldp
	belogs.Debug("insertCerDb(): cerModel.CRLdp:", cerModel.CrldpModel.Crldps)
	if len(cerModel.CrldpModel.Crldps) > 0 {
		sqlStr = `INSERT lab_rpki_cer_crldp(cerId, crldp) VALUES(?,?)`
		for _, crldp := range cerModel.CrldpModel.Crldps {
			res, err = session.Exec(sqlStr, cerId, crldp)
			if err != nil {
				belogs.Error("insertCerDb(): INSERT lab_rpki_cer_crldp Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
				return err
			}
		}
	}

	//lab_rpki_cer_ipaddress
	belogs.Debug("insertCerDb(): cerModel.CerIpAddressModel:", cerModel.CerIpAddressModel)
	sqlStr = `INSERT lab_rpki_cer_ipaddress(cerId,addressFamily, addressPrefix,min,max,
	                rangeStart,rangeEnd,addressPrefixRange) 
	                 VALUES(?,?,?,?,?,
	                 ?,?,?)`
	for _, cerIpAddress := range cerModel.CerIpAddressModel.CerIpAddresses {
		res, err = session.Exec(sqlStr,
			cerId, cerIpAddress.AddressFamily, cerIpAddress.AddressPrefix, cerIpAddress.Min, cerIpAddress.Max,
			cerIpAddress.RangeStart, cerIpAddress.RangeEnd, cerIpAddress.AddressPrefixRange)
		if err != nil {
			belogs.Error("insertCerDb(): INSERT lab_rpki_cer_ipaddress Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_cer_sia
	belogs.Debug("insertCerDb(): cerModel.Sia:", cerModel.SiaModel)
	if len(cerModel.SiaModel.CaRepository) > 0 ||
		len(cerModel.SiaModel.RpkiManifest) > 0 ||
		len(cerModel.SiaModel.RpkiNotify) > 0 ||
		len(cerModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT lab_rpki_cer_sia(cerId, rpkiManifest,rpkiNotify,caRepository,signedObject) VALUES(?,?,?,?,?)`
		res, err = session.Exec(sqlStr, cerId, cerModel.SiaModel.RpkiManifest,
			cerModel.SiaModel.RpkiNotify, cerModel.SiaModel.CaRepository,
			cerModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("insertCerDb(): INSERT lab_rpki_cer_sia Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}
	return nil
}

func getExpireCerDb(now time.Time) (certIdStateModels []CertIdStateModel, err error) {

	certIdStateModels = make([]CertIdStateModel, 0)
	t := convert.Time2String(now)
	sql := `select id, state as stateStr, c.NotAfter as endTime from  lab_rpki_cer c 
			where timestamp(c.NotAfter) < ? order by id `

	err = xormdb.XormEngine.SQL(sql, t).Find(&certIdStateModels)
	if err != nil {
		belogs.Error("getExpireCerDb(): lab_rpki_cer fail:", t, err)
		return nil, err
	}
	belogs.Info("getExpireCerDb(): now t:", t, "  , len(certIdStateModels):", len(certIdStateModels))
	return certIdStateModels, nil
}

func updateCerStateDb(certIdStateModels []CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	sql := `update lab_rpki_cer c set c.state = ? where id = ? `
	for i := range certIdStateModels {
		belogs.Debug("updateCerStateDb():  certIdStateModels[i]:", certIdStateModels[i].Id, certIdStateModels[i].StateStr)
		_, err := session.Exec(sql, certIdStateModels[i].StateStr, certIdStateModels[i].Id)
		if err != nil {
			belogs.Error("updateCerStateDb(): UPDATE lab_rpki_cer fail :", jsonutil.MarshalJson(certIdStateModels[i]), err)
			return xormdb.RollbackAndLogError(session, "updateCerStateDb(): UPDATE lab_rpki_cer fail : certIdStateModels[i]: "+
				jsonutil.MarshalJson(certIdStateModels[i]), err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("updateCerStateDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("updateCerStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Now().Sub(start))

	return nil
}
