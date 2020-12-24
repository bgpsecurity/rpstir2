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
func AddRoas(syncLogFileModels []parsevalidatemodel.SyncLogFileModel) error {
	session, err := xormdb.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	start := time.Now()

	belogs.Debug("AddRoas(): len(syncLogFileModels):", len(syncLogFileModels))
	// insert new mft
	for i := range syncLogFileModels {
		err = insertRoa(session, &syncLogFileModels[i], start)
		if err != nil {
			belogs.Error("AddRoas(): insertRoa fail:", jsonutil.MarshalJson(syncLogFileModels[i]), err)
			return xormdb.RollbackAndLogError(session, "AddRoas(): insertRoa fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	err = UpdateSyncLogFilesJsonAllAndState(session, syncLogFileModels)
	if err != nil {
		belogs.Error("AddRoas(): UpdateSyncLogFilesJsonAllAndState fail:", err)
		return xormdb.RollbackAndLogError(session, "AddRoas(): UpdateSyncLogFilesJsonAllAndState fail ", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("AddRoas(): insertRoa CommitSession fail :", err)
		return err
	}
	belogs.Info("AddRoas(): len(roas):", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}

// del
func DelRoas(delSyncLogFileModels []parsevalidatemodel.SyncLogFileModel, updateSyncLogFileModels []parsevalidatemodel.SyncLogFileModel, wg *sync.WaitGroup) (err error) {
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
	belogs.Debug("DelRoas(): len(syncLogFileModels):", len(syncLogFileModels), jsonutil.MarshalJson(syncLogFileModels))
	for i := range syncLogFileModels {
		err = delRoaById(session, syncLogFileModels[i].CertId)
		if err != nil {
			belogs.Error("DelRoas(): DelRoaById fail, cerId:", syncLogFileModels[i].CertId, err)
			return xormdb.RollbackAndLogError(session, "DelRoas(): DelRoaById fail: "+jsonutil.MarshalJson(syncLogFileModels[i]), err)
		}
	}

	// only update delSyncLogFileModels
	err = UpdateSyncLogFilesJsonAllAndState(session, delSyncLogFileModels)
	if err != nil {
		belogs.Error("DelRoas(): UpdateSyncLogFilesJsonAllAndState fail:", err)
		return xormdb.RollbackAndLogError(session, "DelRoas(): UpdateSyncLogFilesJsonAllAndState fail", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("DelRoas(): CommitSession fail :", err)
		return err
	}
	belogs.Info("DelRoas(): len(roas), ", len(syncLogFileModels), "  time(s):", time.Now().Sub(start).Seconds())
	return nil
}
func DelRoaByFile(session *xorm.Session, filePath, fileName string) (err error) {
	// try to delete old
	belogs.Debug("DelRoaByFile():will delete lab_rpki_roa by filePath+fileName:", filePath, fileName)
	labRpkiRoa := model.LabRpkiRoa{}

	var roaId uint64
	has, err := session.Table(&labRpkiRoa).Where("filePath=?", filePath).And("fileName=?", fileName).Cols("id").Get(&roaId)
	if err != nil {
		belogs.Error("DelRoaByFile(): get current labRpkiRoa fail:", filePath, fileName, err)
		return err
	}

	belogs.Debug("DelRoaByFile():will delete lab_rpki_roa roaId:", roaId, "    has:", has)
	if has {
		return delRoaById(session, roaId)
	}
	return nil
}
func delRoaById(session *xorm.Session, roaId uint64) (err error) {

	belogs.Debug("delRoaById():delete lab_rpki_roa_ by roaId:", roaId)

	// rrdp may have id==0, just return nil
	if roaId <= 0 {
		return nil
	}

	//lab_rpki_roa_ipaddress
	_, err = session.Exec("delete from lab_rpki_roa_ipaddress  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("delRoaById():delete  from lab_rpki_roa_ipaddress fail: roaId: ", roaId, err)
		return err
	}

	//lab_rpki_roa_ee_ipaddress
	_, err = session.Exec("delete from lab_rpki_roa_ee_ipaddress  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("delRoaById():delete  from lab_rpki_roa_ee_ipaddress fail: roaId: ", roaId, err)
		return err
	}

	//lab_rpki_roa_sia
	_, err = session.Exec("delete from  lab_rpki_roa_sia  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("delRoaById():delete  from lab_rpki_roa_sia fail: roaId: ", roaId, err)
		return err
	}

	//lab_rpki_roa_sia
	_, err = session.Exec("delete from  lab_rpki_roa_aia  where roaId = ?", roaId)
	if err != nil {
		belogs.Error("delRoaById():delete  from lab_rpki_roa_aia fail: roaId: ", roaId, err)
		return err
	}

	//lab_rpki_roa
	_, err = session.Exec("delete from  lab_rpki_roa  where id = ?", roaId)
	if err != nil {
		belogs.Error("delRoaById():delete  from lab_rpki_roa fail: roaId: ", roaId, err)
		return err
	}
	return nil
}

func insertRoa(session *xorm.Session,
	syncLogFileModel *parsevalidatemodel.SyncLogFileModel, now time.Time) error {

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
		belogs.Error("insertRoa(): INSERT lab_rpki_roa Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	roaId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("insertRoa(): LastInsertId :", jsonutil.MarshalJson(syncLogFileModel), err)
		return err
	}

	//lab_rpki_roa_aia
	belogs.Debug("insertRoa(): roaModel.Aia.CaIssuers:", roaModel.AiaModel.CaIssuers)
	if len(roaModel.AiaModel.CaIssuers) > 0 {
		sqlStr = `INSERT lab_rpki_roa_aia(roaId, caIssuers)
				VALUES(?,?)`
		res, err = session.Exec(sqlStr, roaId, roaModel.AiaModel.CaIssuers)
		if err != nil {
			belogs.Error("insertRoa(): INSERT lab_rpki_roa_aia Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_roa_sia
	belogs.Debug("insertRoa(): roaModel.Sia:", roaModel.SiaModel)
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
			belogs.Error("insertRoa(): INSERT lab_rpki_roa_sia Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}

	//lab_rpki_roa_ipaddress
	belogs.Debug("insertRoa(): roaModel.IPAddrBlocks:", jsonutil.MarshalJson(roaModel.RoaIpAddressModels))
	if roaModel.RoaIpAddressModels != nil && len(roaModel.RoaIpAddressModels) > 0 {
		sqlStr = `INSERT lab_rpki_roa_ipaddress(roaId, addressFamily,addressPrefix,maxLength, rangeStart, rangeEnd,addressPrefixRange )
						VALUES(?,?,?,?,?,?,?)`
		for _, roaIpAddressModel := range roaModel.RoaIpAddressModels {
			res, err = session.Exec(sqlStr, roaId, roaIpAddressModel.AddressFamily,
				roaIpAddressModel.AddressPrefix, roaIpAddressModel.MaxLength,
				roaIpAddressModel.RangeStart, roaIpAddressModel.RangeEnd, roaIpAddressModel.AddressPrefixRange)
			if err != nil {
				belogs.Error("insertRoa(): INSERT lab_rpki_roa_ipaddress Exec :", jsonutil.MarshalJson(syncLogFileModel), err)
				return err
			}

		}
	}

	//lab_rpki_roa_ee_ipaddress
	belogs.Debug("insertRoa(): roaModel.CerIpAddressModel:", roaModel.EeCertModel.CerIpAddressModel)
	sqlStr = `INSERT lab_rpki_roa_ee_ipaddress(roaId,addressFamily, addressPrefix,min,max,
	                rangeStart,rangeEnd,addressPrefixRange) 
	                 VALUES(?,?,?,?,?,
	                 ?,?,?)`
	for _, cerIpAddress := range roaModel.EeCertModel.CerIpAddressModel.CerIpAddresses {
		res, err = session.Exec(sqlStr,
			roaId, cerIpAddress.AddressFamily, cerIpAddress.AddressPrefix, cerIpAddress.Min, cerIpAddress.Max,
			cerIpAddress.RangeStart, cerIpAddress.RangeEnd, cerIpAddress.AddressPrefixRange)
		if err != nil {
			belogs.Error("insertRoa(): INSERT lab_rpki_roa_ee_ipaddress Exec:", jsonutil.MarshalJson(syncLogFileModel), err)
			return err
		}
	}
	return nil
}
