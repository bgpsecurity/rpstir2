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

func InsertCerDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("InsertCerDb(): will add cer file:", syncLogFileModel.FilePath, syncLogFileModel.FileName,
		"  fileType:", syncLogFileModel.FileType)

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("InsertCerDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()

	err = InsertCerDbWithSession(session, syncLogFileModel, start)
	if err != nil {
		belogs.Error("InsertCerDb(): InsertCerDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertCerDb(): InsertCerDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("InsertCerDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return xormdb.RollbackAndLogError(session, "InsertCerDb(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("InsertCerDb(): CommitSession fail :", err)
		return err
	}
	belogs.Info("InsertCerDb(): cer file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func InsertCerDbWithSession(session *xorm.Session,
	syncLogFileModel *model.SyncLogFileModel, now time.Time) error {
	if syncLogFileModel.CertModel == nil {
		belogs.Error("InsertCerDbWithSession(): CertModel is nil, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is nil")
	}

	var cerModel model.CerModel
	json := jsonutil.MarshalJson(syncLogFileModel.CertModel)
	belogs.Debug("InsertCerDbWithSession():cerModel json:", json)
	err := jsonutil.UnmarshalJson(json, &cerModel)
	if err != nil {
		belogs.Error("InsertCerDbWithSession(): json fail, CertModel to cerModel, syncLogFileModel:", syncLogFileModel.String())
		return errors.New("CertModel is not cerModel type")
	}
	orginModelJson := jsonutil.MarshalJson(syncLogFileModel.OriginModel)
	notBefore := cerModel.NotBefore
	notAfter := cerModel.NotAfter
	belogs.Debug("InsertCerDbWithSession():cerModel filePath,fileName:", cerModel.FilePath, cerModel.FileName, "  orginModel:", orginModelJson,
		"  now ", now, "  notBefore:", notBefore, "  notAfter:", notAfter)

	//lab_rpki_cer
	sqlStr := `INSERT INTO lab_rpki_cer(
	    sn, notBefore, notAfter, subject,
	    issuer, ski, aki, filePath, fileName,
	    fileHash, jsonAll, syncLogId, syncLogFileId, updateTime,
	    state, origin) 	
	    VALUES(?,?,?,?,
	    ?,?,?,?,?,
	    ?,?,?,?,?,
	    ?,?)`
	res, err := session.Exec(sqlStr,
		cerModel.Sn, notBefore, notAfter, cerModel.Subject,
		cerModel.Issuer, xormdb.SqlNullString(cerModel.Ski), xormdb.SqlNullString(cerModel.Aki), cerModel.FilePath, cerModel.FileName,
		cerModel.FileHash, xormdb.SqlNullString(jsonutil.MarshalJson(cerModel)), syncLogFileModel.SyncLogId, syncLogFileModel.Id, now,
		xormdb.SqlNullString(jsonutil.MarshalJson(syncLogFileModel.StateModel)),
		xormdb.SqlNullString(orginModelJson))
	if err != nil {
		belogs.Error("InsertCerDbWithSession(): INSERT INTO lab_rpki_cer fail, ",
			"  cerModel:", cerModel.String(), "     syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}

	cerId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("InsertCerDbWithSession(): LastInsertId fail:", syncLogFileModel.String(), err)
		return err
	}

	//lab_rpki_cer_aia
	belogs.Debug("InsertCerDbWithSession(): cerId:", cerId, " cerModel.Aia.CaIssuers:", cerModel.AiaModel.CaIssuers)
	if len(cerModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT INTO lab_rpki_cer_aia(cerId, caIssuers) VALUES(?,?)`
		_, err = session.Exec(sqlStr, cerId, cerModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("InsertCerDbWithSession(): INSERT INTO lab_rpki_cer_aia Exec:", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_cer_asn
	belogs.Debug("InsertCerDbWithSession(): cerModel.Asn:", cerModel.AsnModel)
	if len(cerModel.AsnModel.Asns) > 0 {
		sqlAsnStr := `INSERT INTO lab_rpki_cer_asn(cerId, asn) VALUES(?,?)`
		sqlMinMaxStr := `INSERT INTO lab_rpki_cer_asn(cerId, min,max) VALUES(?,?,?)`
		for _, asn := range cerModel.AsnModel.Asns {
			// need  asNum >=0
			if asn.Asn >= 0 {
				_, err = session.Exec(sqlAsnStr, cerId, asn.Asn)
				if err != nil {
					belogs.Error("InsertCerDbWithSession(): INSERT sqlAsnStr lab_rpki_cer_asn ,syncLogFileModel err:", syncLogFileModel.String(), err)
					return err
				}
			} else if asn.Max >= 0 && asn.Min >= 0 {
				_, err = session.Exec(sqlMinMaxStr, cerId, asn.Min, asn.Max)
				if err != nil {
					belogs.Error("InsertCerDbWithSession(): INSERT sqlMinMaxStr lab_rpki_cer_asn,syncLogFileModel err:", syncLogFileModel.String(), err)
					return err
				}
			} else {
				belogs.Error("InsertCerDbWithSession(): INSERT INTO lab_rpki_cer_asn asn/min/max all are zero, syncLogFileModel err:", syncLogFileModel.String())
				return errors.New("INSERT INTO lab_rpki_cer_asn fail, asn/min/max all are zero")
			}
		}
	}

	//lab_rpki_cer_crldp
	belogs.Debug("InsertCerDbWithSession(): cerModel.CRLdp:", cerModel.CrldpModel.Crldps)
	if len(cerModel.CrldpModel.Crldps) > 0 {
		sqlStr = `INSERT INTO lab_rpki_cer_crldp(cerId, crldp) VALUES(?,?)`
		for _, crldp := range cerModel.CrldpModel.Crldps {
			_, err = session.Exec(sqlStr, cerId, crldp)
			if err != nil {
				belogs.Error("InsertCerDbWithSession(): INSERT INTO lab_rpki_cer_crldp Exec:", syncLogFileModel.String(), err)
				return err
			}
		}
	}

	//lab_rpki_cer_ipaddress
	belogs.Debug("InsertCerDbWithSession(): cerModel.CerIpAddressModel:", cerModel.CerIpAddressModel)
	sqlStr = `INSERT INTO lab_rpki_cer_ipaddress(cerId,addressFamily, addressPrefix,min,max,
	                rangeStart,rangeEnd,addressPrefixRange) 
	                 VALUES(?,?,?,?,?,
	                 ?,?,?)`
	for _, cerIpAddress := range cerModel.CerIpAddressModel.CerIpAddresses {
		_, err = session.Exec(sqlStr,
			cerId, cerIpAddress.AddressFamily, cerIpAddress.AddressPrefix, cerIpAddress.Min, cerIpAddress.Max,
			cerIpAddress.RangeStart, cerIpAddress.RangeEnd, cerIpAddress.AddressPrefixRange)
		if err != nil {
			belogs.Error("InsertCerDbWithSession(): INSERT INTO lab_rpki_cer_ipaddress Exec:", syncLogFileModel.String(), err)
			return err
		}
	}

	//lab_rpki_cer_sia
	belogs.Debug("InsertCerDbWithSession(): cerModel.Sia:", cerModel.SiaModel)
	if len(cerModel.SiaModel.CaRepository) > 0 ||
		len(cerModel.SiaModel.RpkiManifest) > 0 ||
		len(cerModel.SiaModel.RpkiNotify) > 0 ||
		len(cerModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT INTO lab_rpki_cer_sia(cerId, rpkiManifest,rpkiNotify,caRepository,signedObject) VALUES(?,?,?,?,?)`
		_, err = session.Exec(sqlStr, cerId, cerModel.SiaModel.RpkiManifest,
			cerModel.SiaModel.RpkiNotify, cerModel.SiaModel.CaRepository,
			cerModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("InsertCerDbWithSession(): INSERT INTO lab_rpki_cer_sia Exec:", syncLogFileModel.String(), err)
			return err
		}
	}
	return nil
}

func DelCerDb(syncLogFileModel *model.SyncLogFileModel) (err error) {
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("DelCerDb(): NewSession fail:", err)
		return err
	}
	defer session.Close()
	err = DelCerDbWithSession(session, syncLogFileModel)
	if err != nil {
		belogs.Error("DelCerDb(): DelCerDbWithSession fail:", err)
		return xormdb.RollbackAndLogError(session, "DelCerDb(): DelCerDbWithSession fail, syncLogFileModel:"+syncLogFileModel.String(), err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelCerDb(): CommitSession fail :", err)
		return err
	}
	return nil
}

func DelCerDbWithSession(session *xorm.Session, syncLogFileModel *model.SyncLogFileModel) (err error) {
	start := time.Now()
	belogs.Debug("DelCerDbWithSession(): will del cer file, certId:", syncLogFileModel.CertId,
		" filePath:", syncLogFileModel.FilePath, "  fileName:", syncLogFileModel.FileName)

	if syncLogFileModel.CertId == 0 {
		certId, err := getCertIdByFilePathNameWithSession(session, "lab_rpki_cer", syncLogFileModel.FilePath, syncLogFileModel.FileName)
		if err != nil {
			belogs.Error("DelCerDbWithSession(): getCertIdByFilePathNameWithSession fail, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName, err)
			return err
		}
		if certId == 0 {
			belogs.Info("DelCerDbWithSession(): file not exist in db, just return, filePath:", syncLogFileModel.FilePath,
				"  fileName:", syncLogFileModel.FileName)
			return nil
		}
		syncLogFileModel.CertId = certId
		belogs.Debug("DelCerDbWithSession(): get certId, certId:", syncLogFileModel.CertId)
	}

	err = DelCerByIdDbWithSession(session, syncLogFileModel.CertId)
	if err != nil {
		belogs.Error("DelCerDbWithSession(): DelCerByIdDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
		return err
	}
	// only del,will update syncLogFile.
	// when is add/update, will update syncLogFile in InsertCerDbWithSession()
	if syncLogFileModel.SyncType == "del" {
		err = UpdateSyncLogFileJsonAllAndStateDbWithSession(session, syncLogFileModel)
		if err != nil {
			belogs.Error("DelCerDbWithSession(): UpdateSyncLogFileJsonAllAndStateDbWithSession fail, syncLogFileModel:", syncLogFileModel.String(), err)
			return err
		}
	}

	belogs.Info("DelCerDbWithSession(): cer file:", syncLogFileModel.FilePath, syncLogFileModel.FileName, "  time(s):", time.Since(start))
	return nil
}

func DelCerByIdDbWithSession(session *xorm.Session, cerId uint64) (err error) {

	belogs.Debug("DelCerByIdDbWithSession():delete lab_rpki_cer by cerId:", cerId)
	// rrdp may have id==0, just return nil
	if cerId <= 0 {
		return nil
	}
	belogs.Info("DelCerByIdDbWithSession():delete lab_rpki_cer by cerId, more than 0:", cerId)

	//lab_rpki_cer_sia
	res, err := session.Exec("delete from lab_rpki_cer_sia  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDbWithSession():delete from lab_rpki_cer_sia failed, cerId:", cerId, "    err:", err)
		return err
	}
	count, _ := res.RowsAffected()
	belogs.Debug("DelCerByIdDbWithSession():delete lab_rpki_cer_sia by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_ipaddress
	res, err = session.Exec("delete from  lab_rpki_cer_ipaddress  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDbWithSession():delete  from lab_rpki_cer_ipaddress failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelCerByIdDbWithSession():delete lab_rpki_cer_ipaddress by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_crldp
	res, err = session.Exec("delete  from lab_rpki_cer_crldp  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDbWithSession():delete  from lab_rpki_cer_crldp failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelCerByIdDbWithSession():delete lab_rpki_cer_crldp by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_asn
	res, err = session.Exec("delete  from lab_rpki_cer_asn  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDbWithSession():delete  from lab_rpki_cer_asn  failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelCerByIdDbWithSession():delete lab_rpki_cer_asn by cerId:", cerId, "  count:", count)

	//lab_rpki_cer_aia
	res, err = session.Exec("delete  from lab_rpki_cer_aia  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDbWithSession():delete  from lab_rpki_cer_aia  failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelCerByIdDbWithSession():delete lab_rpki_cer_aia by cerId:", cerId, "  count:", count)

	//lab_rpki_cer
	res, err = session.Exec("delete  from lab_rpki_cer  where id = ?", cerId)
	if err != nil {
		belogs.Error("DelCerByIdDbWithSession():delete  from lab_rpki_cer  failed, cerId:", cerId, err)
		return err
	}
	count, _ = res.RowsAffected()
	belogs.Debug("DelCerByIdDbWithSession():delete lab_rpki_cer by cerId:", cerId, "  count:", count)

	return nil
}

/*
func getExpireCerDb(now time.Time) (certIdStateModels []model.CertIdStateModel, err error) {

	certIdStateModels = make([]model.CertIdStateModel, 0)
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
*/

/*
func updateCerStateDb(certIdStateModels []model.CertIdStateModel) error {
	start := time.Now()
	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("updateCerStateDb(): NewSession fail:", err)
		return err
	}
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
	belogs.Info("updateCerStateDb(): len(certIdStateModels):", len(certIdStateModels), "  time(s):", time.Since(start))

	return nil
}
*/
/*
func UpdateCerByCheckAll(now time.Time) error {
	// check expire
	curCertIdStateModels, err := getExpireCerDb(now)
	if err != nil {
		belogs.Error("UpdateCerByCheckAll(): getExpireCerDb:  err: ", err)
		return err
	}
	belogs.Info("UpdateCerByCheckAll(): len(curCertIdStateModels):", len(curCertIdStateModels))

	newCertIdStateModels := make([]model.CertIdStateModel, 0)
	for i := range curCertIdStateModels {
		// if have this error, ignore
		belogs.Debug("UpdateCerByCheckAll(): old curCertIdStateModels[i]:", jsonutil.MarshalJson(curCertIdStateModels[i]))
		if strings.Contains(curCertIdStateModels[i].StateStr, "NotAfter is earlier than the current time") {
			continue
		}

		// will add error
		stateModel := model.StateModel{}
		jsonutil.UnmarshalJson(curCertIdStateModels[i].StateStr, &stateModel)
		stateMsg := model.StateMsg{Stage: "parsevalidate",
			Fail:   "NotAfter is earlier than the current time",
			Detail: "The current time is " + convert.Time2StringZone(now) + ", notAfter is " + convert.Time2StringZone(curCertIdStateModels[i].EndTime)}
		if conf.Bool("policy::allowStaleCer") {
			stateModel.AddWarning(&stateMsg)
		} else {
			stateModel.AddError(&stateMsg)
		}

		certIdStateModel := model.CertIdStateModel{
			Id:       curCertIdStateModels[i].Id,
			StateStr: jsonutil.MarshalJson(stateModel),
		}
		newCertIdStateModels = append(newCertIdStateModels, certIdStateModel)
		belogs.Debug("UpdateCerByCheckAll(): new certIdStateModel:", jsonutil.MarshalJson(certIdStateModel))
	}

	// update db
	err = updateCerStateDb(newCertIdStateModels)
	if err != nil {
		belogs.Error("UpdateCerByCheckAll(): updateCerStateDb:  err: ", len(newCertIdStateModels), err)
		return err
	}
	belogs.Info("UpdateCerByCheckAll(): ok len(newCertIdStateModels):", len(newCertIdStateModels))
	return nil
}
*/
