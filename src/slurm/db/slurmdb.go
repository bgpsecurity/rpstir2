package slurm

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	belogs "github.com/astaxie/beego/logs"
	iputil "github.com/cpusoft/goutil/iputil"
	jsonutil "github.com/cpusoft/goutil/jsonutil"
	osutil "github.com/cpusoft/goutil/osutil"
	xormdb "github.com/cpusoft/goutil/xormdb"

	"model"
)

//
func SaveSlurm(slurm *model.Slurm, slurmFile string) (err error) {
	session, err := xormdb.NewSession()
	defer session.Close()

	filePath, fileName := osutil.Split(slurmFile)

	//lab_rpki_slurm_file
	sqlStr := `INSERT lab_rpki_slurm_file(jsonAll, uploadTime, filePath,fileName) VALUES (?,?,?,?)`
	res, err := session.Exec(sqlStr, jsonutil.MarshalJson(slurm), time.Now(), filePath, fileName)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "SaveSlurm(): INSERT lab_rpki_slurm fail:"+slurmFile, err)
	}

	slurmFileId, err := res.LastInsertId()
	if err != nil {
		belogs.Error("SaveSlurm(): LastInsertId fail:", slurmFile, err)
		return err
	}
	belogs.Debug("SaveSlurm():LastInsertId slurmFileId:", slurmFileId)

	//lab_rpki_slurm
	stat := `{"rtr": "notYet"}`
	sqlStr = `INSERT ignore into lab_rpki_slurm(version,style, asn, addressPrefix, maxLength,ski, routerPublicKey,comment, slurmFileId,state) 
	  VALUES (?,?,?,?,  ?,?,?,?,  ?,?)`
	for _, one := range slurm.LocallyAddedAssertions.PrefixAssertions {
		if one.MaxPrefixLength == 0 {
			_, one.MaxPrefixLength, _ = iputil.SplitAddressAndPrefix(one.Prefix)
		}
		belogs.Debug("SaveSlurm(): INSERT lab_rpki_slurm prefixAssertions:", slurm.SlurmVersion, jsonutil.MarshalJson(one), slurmFileId)
		_, err := session.Exec(sqlStr, slurm.SlurmVersion, "prefixAssertions", one.Asn.SqlNullInt(), xormdb.SqlNullString(one.Prefix),
			xormdb.SqlNullInt(int64(one.MaxPrefixLength)), sql.NullString{}, sql.NullString{}, xormdb.SqlNullString(one.Comment), slurmFileId, stat)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "SaveSlurm(): INSERT lab_rpki_slurm prefixAssertions fail:"+slurmFile, err)
		}
	}
	for _, one := range slurm.LocallyAddedAssertions.BgpsecAssertions {
		belogs.Debug("SaveSlurm(): INSERT lab_rpki_slurm bgpsecAssertions:", slurm.SlurmVersion, jsonutil.MarshalJson(one), slurmFileId)
		_, err := session.Exec(sqlStr, slurm.SlurmVersion, "bgpsecAssertions", one.Asn.SqlNullInt(), sql.NullString{},
			sql.NullInt64{}, xormdb.SqlNullString(one.SKI), xormdb.SqlNullString(one.RouterPublicKey), xormdb.SqlNullString(one.Comment), slurmFileId, stat)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "SaveSlurm(): INSERT lab_rpki_slurm bgpsecAssertions fail:"+slurmFile, err)
		}
	}

	for _, one := range slurm.ValidationOutputFilters.PrefixFilters {
		belogs.Debug("SaveSlurm(): INSERT lab_rpki_slurm prefixFilters:", slurm.SlurmVersion, jsonutil.MarshalJson(one), slurmFileId)
		_, err := session.Exec(sqlStr, slurm.SlurmVersion, "prefixFilters", one.Asn.SqlNullInt(), xormdb.SqlNullString(one.Prefix),
			sql.NullInt64{}, sql.NullString{}, sql.NullString{}, xormdb.SqlNullString(one.Comment), slurmFileId, stat)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "SaveSlurm(): INSERT lab_rpki_slurm prefixFilters fail:"+slurmFile, err)
		}
	}
	for _, one := range slurm.ValidationOutputFilters.BgpsecFilters {
		belogs.Debug("SaveSlurm(): INSERT lab_rpki_slurm bgpsecFilters:", slurm.SlurmVersion, jsonutil.MarshalJson(one), slurmFileId)
		_, err := session.Exec(sqlStr, slurm.SlurmVersion, "bgpsecFilters", one.Asn.SqlNullInt(), sql.NullString{},
			sql.NullInt64{}, xormdb.SqlNullString(one.SKI), sql.NullString{}, xormdb.SqlNullString(one.Comment), slurmFileId, stat)
		if err != nil {
			return xormdb.RollbackAndLogError(session, "SaveSlurm(): INSERT lab_rpki_slurm bgpsecFilters fail:"+slurmFile, err)
		}
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("SaveSlurm(): CommitSession lab_rpki_sync_log :", err)
		return err
	}
	return nil
}

func CheckConflictInDb(style string, asn model.SlurmAsnModel,
	prefix string, maxPrefixLength uint64) (err error) {
	belogs.Debug("CheckConflictInDb():style, asn, prefix, maxPrefixLength:", style, asn, prefix, maxPrefixLength)

	var styleDb string
	if style == "prefixAssertions" {
		styleDb = "prefixFilters"
	} else if style == "prefixFilters" {
		styleDb = "prefixAssertions"
	} else {
		belogs.Error("CheckConflictInDb(): style is error :", style)
		return err
	}
	labRpkiSlurm := new(model.LabRpkiSlurm)

	s := xormdb.XormEngine.Where("style = ?", styleDb)
	if asn.IsNotNil {
		s = s.And("asn=?", asn.Value)
	}
	s = s.And("addressPrefix=?", prefix).And("maxLength=?", maxPrefixLength)

	total, err := s.Count(labRpkiSlurm)
	if err != nil {
		belogs.Error("CheckConflictInDb():Count fail,style, asn,prefix, maxPrefixLength,: ",
			style, asn, prefix, maxPrefixLength, err)
		return err
	}
	belogs.Debug("CheckConflictInDb():Count, total :", total)
	if total > 0 {
		belogs.Error("CheckConflictInDb():Count is not zero, style, asn,prefix, maxPrefixLength,: ",
			style, asn, prefix, maxPrefixLength)
		return errors.New("find conflict with with existing data, one item is '" + style + ", " + strconv.Itoa(int(asn.Value)) +
			", " + prefix + ", " + strconv.Itoa(int(maxPrefixLength)) + "'")
	}

	return nil
}
