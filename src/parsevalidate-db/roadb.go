package parsevalidatedb

import (
	"errors"
	"time"

	"github.com/bgpsecurity/rpstir2/model"
	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

func InsertRoaDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("InsertRoaDb(): will add roa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("InsertRoaDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = InsertRoaDbWithSession(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("InsertRoaDb(): InsertRoaDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertRoaDb(): InsertRoaDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("InsertRoaDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertRoaDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("InsertRoaDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("InsertRoaDb(): roa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func InsertRoaDbWithSession(session *xorm.Session,
	syncLogFileModel *model.SyncLogFileModel, now time.Time) error {

	if syncLogFileModel.CertModel == nil {
		belogs.Error("InsertRoaDbWithSession(): CertModel is nil, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is nil")
	}

	var roaModel model.RoaModel
	json := jsonutil.MarshalJson(syncLogFileModel.CertModel)
	belogs.Debug("InsertRoaDbWithSession():roaModel json:", json)
	err := jsonutil.UnmarshalJson(json, &roaModel)
	if err != nil {
		belogs.Error("InsertRoaDbWithSession(): json fail, CertModel to RoaModel, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is not roaModel type")
	}
	originModelJson := jsonutil.MarshalJson(syncLogFileModel.OriginModel)
	belogs.Debug("InsertRoaDbWithSession(): roaModel filePath,fileName:", roaModel.FilePath, roaModel.FileName, " originModel:", originModelJson, " now ", now)

	//lab_rpki_roa
	sqlStr := `INSERT INTO lab_rpki_roa(
	                asn, ski, aki, filePath, fileName, 
	                fileHash, jsonAll, syncLogId, syncLogFileId, updateTime,
	                state, origin)
					VALUES(?,?,?,?,?,
					?,?,?,?,?,
					?,?)`
	res, err := session.Exec(sqlStr,
		roaModel.Asn, xormdb.SqlNullString(roaModel.Ski), xormdb.SqlNullString(roaModel.Aki), roaModel.FilePath, roaModel.FileName,
		roaModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(roaModel)), syncLogFileModel.SyncLogId, syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)),
		xormdb.SqlNullString(originModelJson))
	if err != nil {
		belogs.Error("InsertRoaDbWithSession(): INSERT INTO lab_rpki_roa Exec fail:",
			"  roaModel:", roaModel.String(), "   syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}

	roaId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("InsertRoaDbWithSession(): LastInsertId fail:", syncLogFileModel.String(), err)
		return err
	}

	//lab_rpki_roa_aia
	belogs.Debug("InsertRoaDbWithSession(): roaId:", roaId, " roaModel.Aia.CaIssuers:", roaModel.AiaModel.CaIssuers)
	if len(roaModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT INTO lab_rpki_roa_aia(roaId, caIssuers)
				VALUES(?,?)`
		_, err = session.Exec(sqlStr, roaId, roaModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("InsertRoaDbWithSession(): INSERT INTO lab_rpki_roa_aia Exec :", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_roa_sia
	belogs.Debug("InsertRoaDbWithSession(): roaModel.Sia:", roaModel.SiaModel)
	if len(roaModel.SiaModel.CaRepository) > 0 ||
		len(roaModel.SiaModel.RpkiManifest) > 0 ||
		len(roaModel.SiaModel.RpkiNotify) > 0 ||
		len(roaModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT INTO lab_rpki_roa_sia(roaId, rpkiManifest,rpkiNotify,caRepository,signedObject)
				VALUES(?,?,?,?,?)`
		_, err = session.Exec(sqlStr, roaId, roaModel.SiaModel.RpkiManifest,
			roaModel.SiaModel.RpkiNotify, roaModel.SiaModel.CaRepository,
			roaModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("InsertRoaDbWithSession(): INSERT INTO lab_rpki_roa_sia Exec :", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_roa_ipaddress
	belogs.Debug("InsertRoaDbWithSession(): roaModel.IPAddrBlocks:", jsonutil.MarshalJson(roaModel.RoaIpAddressModels))
	if roaModel.RoaIpAddressModels != nil && len(roaModel.RoaIpAddressModels) > 0 {
		sqlStr = `INSERT INTO lab_rpki_roa_ipaddress(roaId, addressFamily,addressPrefix,maxLength, rangeStart, rangeEnd,addressPrefixRange )
						VALUES(?,?,?,?,?,?,?)`
		for _, roaIpAddressModel := range roaModel.RoaIpAddressModels {
			_, err = session.Exec(sqlStr, roaId, roaIpAddressModel.AddressFamily,
				roaIpAddressModel.AddressPrefix, roaIpAddressModel.MaxLength,
				roaIpAddressModel.RangeStart, roaIpAddressModel.RangeEnd, roaIpAddressModel.AddressPrefixRange)
			if err != nil {
				belogs.Error("InsertRoaDbWithSession(): INSERT INTO lab_rpki_roa_ipaddress Exec :", syncLogFileModel.String(), err)
				return err
			}

		}
	}

	//lab_rpki_roa_ee_ipaddress
	belogs.Debug("InsertRoaDbWithSession(): roaModel.CerIpAddressModel:", roaModel.EeCertModel.CerIpAddressModel)
	sqlStr = `INSERT INTO lab_rpki_roa_ee_ipaddress(roaId,addressFamily, addressPrefix,min,max,
	                rangeStart,rangeEnd,addressPrefixRange) 
	                 VALUES(?,?,?,?,?,
	                 ?,?,?)`
	for _, cerIpAddress := range roaModel.EeCertModel.CerIpAddressModel.CerIpAddresses {
		_, err = session.Exec(sqlStr,
			roaId, cerIpAddress.AddressFamily, cerIpAddress.AddressPrefix, cerIpAddress.Min, cerIpAddress.Max,
			cerIpAddress.RangeStart, cerIpAddress.RangeEnd, cerIpAddress.AddressPrefixRange)
		if err != nil {
			belogs.Error("InsertRoaDbWithSession(): INSERT INTO lab_rpki_roa_ee_ipaddress Exec:", syncLogFileModel.String(), err)
			return err
		}
	}
	return nil
}

func DelRoaDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("DelRoaDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	err = DelRoaDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("DelRoaDb(): DelRoaDbWithSession fail:", err)
		return xormdb.RollbackAndLogError(session, "DelRoaDb(): DelRoaDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelRoaDb(): CommitSession fail :", err)
		return err
	}
	return nil
}

func DelRoaDbWithSession(session *xorm.Session, syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("DelRoaDbWithSession(): will del roa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName)

	if syncLogFileModel.CertId == 0 {
		certId, err := getCertIdByFilePathNameWithSession(session, "lab_rpki_roa", syncLogFileModel.FilePath, syncLogFileModel.FileName)
		if err != nil {
			belogs.Error("DelRoaDbWithSession(): getCertIdByFilePathNameWithSession fail, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName, err)
			return err
		}
		if certId == 0 {
			belogs.Info("DelRoaDbWithSession(): file not exist in db, just return, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName)
			return nil
		}
		syncLogFileModel.CertId = certId
		belogs.Debug("DelRoaDbWithSession(): get certId, certId:", syncLogFileModel.CertId)
	}

	err = DelRoaByIdDbWithSession(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("DelRoaDbWithSession(): DelRoaByIdDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}
	// only del,will update syncLogFile.
	// when is add/update, will update syncLogFile in InsertRoaDbWithSession()
	if syncLogFileModel.SyncType == "del" {
		err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
		if err != nil {
			belogs.Error("DelRoaDbWithSession(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}
	}

	belogs.Info("DelRoaDbWithSession(): roa file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelRoaByIdDbWithSession(session *xorm.Session, roaId uint64) (err error) {

	belogs.Debug("DelRoaByIdDbWithSession():delete lab_rpki_roa by roaId:", roaId)
	// rrdp may have id==0, just return nil
	if roaId <= 0 {
		return nil
	}
	belogs.Info("DelRoaByIdDbWithSession():delete lab_rpki_roa by roaId, more than 0:", roaId)

	//lab_rpki_roa_ipaddress
	res, err := session.Exec("delete from lab_rpki_roa_ipaddress  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("DelRoaByIdDbWithSession():delete  from lab_rpki_roa_ipaddress fail: roaId: ", roaId, err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("DelRoaByIdDbWithSession():delete lab_rpki_roa_ipaddress by roaId:", roaId, "  count:", count)

	//lab_rpki_roa_ee_ipaddress
	res, err = session.Exec("delete from lab_rpki_roa_ee_ipaddress  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("DelRoaByIdDbWithSession():delete  from lab_rpki_roa_ee_ipaddress fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelRoaByIdDbWithSession():delete lab_rpki_roa_ee_ipaddress by roaId:", roaId, "  count:", count)

	//lab_rpki_roa_sia
	res, err = session.Exec("delete from  lab_rpki_roa_sia  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("DelRoaByIdDbWithSession():delete  from lab_rpki_roa_sia fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelRoaByIdDbWithSession():delete lab_rpki_roa_sia by roaId:", roaId, "  count:", count)

	//lab_rpki_roa_aia
	res, err = session.Exec("delete from  lab_rpki_roa_aia  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("DelRoaByIdDbWithSession():delete  from lab_rpki_roa_aia fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelRoaByIdDbWithSession():delete lab_rpki_roa_aia by roaId:", roaId, "  count:", count)

	//lab_rpki_roa
	res, err = session.Exec("delete from  lab_rpki_roa  where id = ?", roaId)
	if err != nil {
		belogs.Error("DelRoaByIdDbWithSession():delete  from lab_rpki_roa fail: roaId: ", roaId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelRoaByIdDbWithSession():delete lab_rpki_roa by roaId:", roaId, "  count:", count)

	return nil
}

/*
func getExpireRoaDb(now time.Time) (certIdStateModels []model.CertIdStateModel, err error) {

	certIdStateModels = make([]model.CertIdStateModel, 0)
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

func updateRoaStateDb(certIdStateModels []model.CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateRoaStateDb(): NewSession fail:", err)
		return err
	}
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
	belogs.Info("updateRoaStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Since(start))

	return nil
}

func UpdateRoaByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireRoaDb(now)
	if err != nil {
		belogs.Error("UpdateRoaByCheckAll(): getExpireRoaDb:  err: ", err)
		return err
	}
	belogs.Info("UpdateRoaByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]model.CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("UpdateRoaByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
		if strings.Contains(curCertIdStateModels[i].StateStr, "NotAfter of EE is earlier than the current time") {
			continue
		}

		// will add error
		stateModel := model.StateModel{}
		jsonutil.UnmarshalJson(curCertIdStateModels[i].StateStr, &stateModel)

		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NotAfter of EE is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", notAfter is " + convert.Time2StringZone(curCertIdStateModels[i].EndTime)}
		if conf.Bool("policy::allowStaleEe") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}

		certIdStateModel := model.CertIdStateModel{
			Id:       curCertIdStateModels[i].Id,
			StateStr: jsonutil.MarshalJson(stateModel),
		}
		newCertIdStateModels = append(newCertIdStateModels, certIdStateModel)
		belogs.Debug("UpdateRoaByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateRoaStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("UpdateRoaByCheckAll(): updateRoaStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("UpdateRoaByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil

}
*/
