package db

import (
	"errors"
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
func AddCers(syncLogFileModels []parsevalidatemodel.SyncLogFileModel) error {
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	start := time.Now()

	belogs.Debug("AddCers(): len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		// insert new cer
		err = insertCer(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("AddCers(): insertCer fail:", jsonutil.MarshalJson(&syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "AddCers(): insertCer fail: "+jsonutil.MarshalJson(&syncLogFileModels[i]), err)
		}
	}

	err = UpdateSyncLogFilesJsonAllAndState(session, syncLogFileModels)
	if err != nil {
		belogs.Error("AddCers(): UpdateSyncLogFilesJsonAllAndState fail:", err)
		return xormdb.RollbackAndLogError(session, "AddCers(): UpdateSyncLogFilesJsonAllAndState fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("AddCers(): insertCer CommitSession fail :", err)
		return err
	}
	belogs.Info("AddCers(): len(cers):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil

}

// del
func DelCers(delSyncLogFileModels []parsevalidatemodel.SyncLogFileModel, updateSyncLogFileModels []parsevalidatemodel.SyncLogFileModel, wg *sync.WaitGroup) (err error) {
	defer func() {
		wg.Done()
	}()

	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()

	syncLogFileModels := append(delSyncLogFileModels, updateSyncLogFileModels...)
	belogs.Debug("DelCers(): len(syncLogFileModels):", len(syncLogFileModels))
	for i := range syncLogFileModels {
		err = delCerById(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("DelCers(): DelCerById fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "DelCers(): DelCerById fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = UpdateSyncLogFilesJsonAllAndState(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("DelCers(): UpdateSyncLogFilesJsonAllAndState fail:", err)
		return xormdb.RollbackAndLogError(session, "DelCers(): UpdateSyncLogFilesJsonAllAndState fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelCers(): CommitSession fail :", err)
		return err
	}
	belogs.Info("DelCers(): len(cers):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

func DelCerByFile(session *xorm.Session, filePath, fileName string) (err error) {
	belogs.Debug("DelCerByFile():will delete lab_rpki_cer by filePath+fileName:", filePath, fileName)
	labRpkiCer := model.LabRpkiCer{}

	var cerId uint64
	has, err := session.Table(&labRpkiCer).Where("filePath=?", filePath).And("fileName=?", fileName).Cols("id").Get(&cerId)
	if err != nil {
		belogs.Error("DelCerByFile(): get current labRpkiCer fail:", filePath, fileName, err)
		return err
	}
	belogs.Debug("DelCerByFile():will delete lab_rpki_cer by cerId:", cerId, "    has:", has)

	if has {
		return delCerById(session, cerId)
	}
	return nil
}
func delCerById(session *xorm.Session, cerId uint64) (err error) {

	belogs.Debug("delCerById():delete lab_rpki_cer_ by cerId:", cerId)
	// rrdp may have id==0, just return nil
	if cerId <= 0 {
		return nil
	}

	//lab_rpki_cer_sia
	_, err = session.Exec("delete from lab_rpki_cer_sia  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("delCerById():delete  from lab_rpki_cer_sia failed, cerId:", cerId, "    err:", err)
		return err
	}

	//lab_rpki_cer_ipaddress
	_, err = session.Exec("delete from  lab_rpki_cer_ipaddress  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("delCerById():delete  from lab_rpki_cer_ipaddress failed, cerId:", cerId, err)
		return err
	}

	//lab_rpki_cer_crldp
	_, err = session.Exec("delete  from lab_rpki_cer_crldp  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("delCerById():delete  from lab_rpki_cer_crldp failed, cerId:", cerId, err)
		return err
	}

	//lab_rpki_cer_asn
	_, err = session.Exec("delete  from lab_rpki_cer_asn  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("delCerById():delete  from lab_rpki_cer_asn  failed, cerId:", cerId, err)
		return err
	}

	//lab_rpki_cer_aia
	_, err = session.Exec("delete  from lab_rpki_cer_aia  where cerId = ?", cerId)
	if err != nil {
		belogs.Error("delCerById():delete  from lab_rpki_cer_aia  failed, cerId:", cerId, err)
		return err
	}

	//lab_rpki_cer
	_, err = session.Exec("delete  from lab_rpki_cer  where id = ?", cerId)
	if err != nil {
		belogs.Error("delCerById():delete  from lab_rpki_cer  failed, cerId:", cerId, err)
		return err
	}

	return nil
}

func insertCer(session *xorm.Session,
	syncLogFileModel *parsevalidatemodel.SyncLogFileModel, now time.Time) error {

	cerModel := syncLogFileModel.CertModel.(model.CerModel)
	notBefore := cerModel.NotBefore
	notAfter := cerModel.NotAfter
	belogs.Debug("insertCer():now ", now, "  notBefore:", notBefore, "  notAfter:", notAfter,
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
		belogs.Error("insertCer(): INSERT lab_rpki_cer Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	cerId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("insertCer(): LastInsertId:", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}
	belogs.Debug("insertCer():LastInsertId cerId:", cerId)

	//lab_rpki_cer_aia
	belogs.Debug("insertCer(): cerModel.Aia.CaIssuers:", cerModel.AiaModel.CaIssuers)
	if len(cerModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT lab_rpki_cer_aia(cerId, caIssuers) VALUES(?,?)`
		res, err = session.Exec(sqlStr, cerId, cerModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("insertCer(): INSERT lab_rpki_cer_aia Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_cer_asn
	belogs.Debug("insertCer(): cerModel.Asn:", cerModel.AsnModel)
	if len(cerModel.AsnModel.Asns) > 0 {
		sqlAsnStr := `INSERT lab_rpki_cer_asn(cerId, asn) VALUES(?,?)`
		sqlMinMaxStr := `INSERT lab_rpki_cer_asn(cerId, min,max) VALUES(?,?,?)`
		for _, asn := range cerModel.AsnModel.Asns {
			// need  asNum >=0
			if asn.Asn >= 0 {
				res, err = session.Exec(sqlAsnStr, cerId, asn.Asn)
				if err != nil {
					belogs.Error("insertCer(): INSERT sqlAsnStr lab_rpki_cer_asn ,syncLogFileModel err:", jsonutil.MarshalJson(syncLogFileModel), err)
					return err
				}
			} else if asn.Max >= 0 && asn.Min >= 0 {
				res, err = session.Exec(sqlMinMaxStr, cerId, asn.Min, asn.Max)
				if err != nil {
					belogs.Error("insertCer(): INSERT sqlMinMaxStr lab_rpki_cer_asn,syncLogFileModel err:", jsonutil.MarshalJson(syncLogFileModel), err)
					return err
				}
			} else {
				belogs.Error("insertCer(): INSERT lab_rpki_cer_asn asn/min/max all are zero, syncLogFileModel err:", jsonutil.MarshalJson(syncLogFileModel))
				return errors.New("insert lab_rpki_cer_asn fail, asn/min/max all are zero")
			}
		}
	}

	//lab_rpki_cer_crldp
	belogs.Debug("insertCer(): cerModel.CRLdp:", cerModel.CrldpModel.Crldps)
	if len(cerModel.CrldpModel.Crldps) > 0 {
		sqlStr = `INSERT lab_rpki_cer_crldp(cerId, crldp) VALUES(?,?)`
		for _, crldp := range cerModel.CrldpModel.Crldps {
			res, err = session.Exec(sqlStr, cerId, crldp)
			if err != nil {
				belogs.Error("insertCer(): INSERT lab_rpki_cer_crldp Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
				return err
			}
		}
	}

	//lab_rpki_cer_ipaddress
	belogs.Debug("insertCer(): cerModel.CerIpAddressModel:", cerModel.CerIpAddressModel)
	sqlStr = `INSERT lab_rpki_cer_ipaddress(cerId,addressFamily, addressPrefix,min,max,
	                rangeStart,rangeEnd,addressPrefixRange) 
	                 VALUES(?,?,?,?,?,
	                 ?,?,?)`
	for _, cerIpAddress := range cerModel.CerIpAddressModel.CerIpAddresses {
		res, err = session.Exec(sqlStr,
			cerId, cerIpAddress.AddressFamily, cerIpAddress.AddressPrefix, cerIpAddress.Min, cerIpAddress.Max,
			cerIpAddress.RangeStart, cerIpAddress.RangeEnd, cerIpAddress.AddressPrefixRange)
		if err != nil {
			belogs.Error("insertCer(): INSERT lab_rpki_cer_ipaddress Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_cer_sia
	belogs.Debug("insertCer(): cerModel.Sia:", cerModel.SiaModel)
	if len(cerModel.SiaModel.CaRepository) > 0 ||
		len(cerModel.SiaModel.RpkiManifest) > 0 ||
		len(cerModel.SiaModel.RpkiNotify) > 0 ||
		len(cerModel.SiaModel.SignedObject) > 0 {
		sqlStr = `INSERT lab_rpki_cer_sia(cerId, rpkiManifest,rpkiNotify,caRepository,signedObject) VALUES(?,?,?,?,?)`
		res, err = session.Exec(sqlStr, cerId, cerModel.SiaModel.RpkiManifest,
			cerModel.SiaModel.RpkiNotify, cerModel.SiaModel.CaRepository,
			cerModel.SiaModel.SignedObject)
		if err != nil {
			belogs.Error("insertCer(): INSERT lab_rpki_cer_sia Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}
	return nil
}
