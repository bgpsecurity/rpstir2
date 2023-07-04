package common

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/iputil"
	"github.com/cpusoft/goutil/jsonutil"
	"github.com/cpusoft/goutil/xormdb"
	model "rpstir2-model"
	"xorm.io/xorm"
)

func GetSerialNumberCountDb() (myCount uint64, err error) {
	start := time.Now()
	sql := `select count(*) as myCount from lab_rpki_rtr_serial_number`
	has, err := xormdb.XormEngine.SQL(sql).Get(&myCount)
	if err != nil {
		belogs.Error("GetSerialNumberCountDb():select count from lab_rpki_rtr_serial_number, fail:", err)
		return 0, err
	}
	if !has {
		belogs.Error("GetSerialNumberCountDb():select count from lab_rpki_rtr_serial_number, !has:")
		return 0, errors.New("has no serialNumber")
	}
	belogs.Info("GetSerialNumberCountDb(): myCount: ", myCount, "  time(s):", time.Since(start))
	return myCount, nil
}

func GetSerialNumberDb() (serialNumberModel *SerialNumberModel, err error) {
	serialNumberModel = new(SerialNumberModel)
	sql := `select serialNumber, globalSerialNumber, subpartSerialNumber from lab_rpki_rtr_serial_number order by id desc limit 1`
	has, err := xormdb.XormEngine.SQL(sql).Get(serialNumberModel)
	if err != nil {
		belogs.Error("GetSerialNumberDb():select serialNumber from lab_rpki_rtr_serial_number order by id desc limit 1 fail:", err)
		return nil, err
	}
	if !has {
		// init serialNumber
		serialNumberModel.SerialNumber = 1
		serialNumberModel.GlobalSerialNumber = 1
		serialNumberModel.SubpartSerialNumber = 1
	}
	belogs.Debug("GetSerialNumberDb():select max(serialNumberModel) lab_rpki_rtr_serial_number, serialNumberModel :", jsonutil.MarshalJson(serialNumberModel))
	return serialNumberModel, nil
}

// tableName:lab_rpki_rtr_full_log / lab_rpki_rtr_full
// getEffectSlurm: for filter to get effected slurm to incremental
func UpdateRtrFullOrFullLogFromSlurmDb(tableName string, newSerialNumber uint64,
	slurmToRtrFullLogs []model.SlurmToRtrFullLog, getEffectSlurm bool) (effectSlurmToRtrFullLogs []model.SlurmToRtrFullLog, err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()
	effectSlurmToRtrFullLogs = make([]model.SlurmToRtrFullLog, 0)
	// insert into rtr_full_log/lab_rpki_rtr_full
	sqlInsertSlurm := `insert  ignore into ` + tableName + `
				(serialNumber,asn,address,prefixLength, maxLength,sourceFrom) values
				(?,?,?,  ?,?,?)`
	belogs.Debug("UpdateRtrFullOrFullLogFromSlurmDb(): will insert/del lab_rpki_rtr_full_log/lab_rpki_rtr_full from slurmToRtrFullLogs,sqlInsertSlurm:", sqlInsertSlurm,
		" newSerialNumber:", newSerialNumber, "    len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))
	for i := range slurmToRtrFullLogs {
		var address string
		sourceFrom := model.LabRpkiRtrSourceFrom{
			Source:         "slurm",
			SlurmId:        slurmToRtrFullLogs[i].SlurmId,
			SlurmLogId:     slurmToRtrFullLogs[i].SlurmLogId,
			SlurmLogFileId: slurmToRtrFullLogs[i].SlurmLogFileId,
		}
		sourceFormJson := jsonutil.MarshalJson(sourceFrom)

		if slurmToRtrFullLogs[i].Style == "prefixAssertions" {
			asn := slurmToRtrFullLogs[i].Asn
			address, _ = iputil.TrimAddressPrefixZero(slurmToRtrFullLogs[i].Address, iputil.GetIpType(slurmToRtrFullLogs[i].Address))
			prefixLength := slurmToRtrFullLogs[i].PrefixLength
			maxLength := slurmToRtrFullLogs[i].MaxLength
			if maxLength == 0 {
				maxLength = slurmToRtrFullLogs[i].PrefixLength
			}

			affected, err := session.Exec(sqlInsertSlurm,
				newSerialNumber,
				asn,
				address,
				prefixLength,
				maxLength,
				sourceFormJson)

			if err != nil {
				belogs.Error("UpdateRtrFullOrFullLogFromSlurmDb():insert into lab_rpki_rtr_full_log/lab_rpki_rtr_full from slurm fail:",
					"  newSerialNumber:", newSerialNumber, "  asn:", asn, "  address:", address,
					"  prefixLength:", prefixLength, "  maxLength:", maxLength, "  sourceFormJson:", sourceFormJson,
					"  affected:", affected, err)
				return nil, xormdb.RollbackAndLogError(session, "UpdateRtrFullOrFullLogFromSlurmDb(): :insert into lab_rpki_rtr_full_log/lab_rpki_rtr_full from slurm fail: ", err)
			}
			addRows, _ := affected.RowsAffected()
			if getEffectSlurm {
				slurmToRtrFullLogs[i].SourceFromJson = sourceFormJson
				effectSlurmToRtrFullLogs = append(effectSlurmToRtrFullLogs, slurmToRtrFullLogs[i])
				belogs.Debug("UpdateRtrFullOrFullLogFromSlurmDb(): getEffectSlurm, slurmToRtrFullLogs:", jsonutil.MarshalJson(slurmToRtrFullLogs[i]))
			}
			belogs.Debug("UpdateRtrFullOrFullLogFromSlurmDb(): insert lab_rpki_rtr_full_log/lab_rpki_rtr_full from slurmToRtrFullLogs, newSerialNumber:",
				newSerialNumber, ",  slurmToRtrFullLogs[i]: ", jsonutil.MarshalJson(slurmToRtrFullLogs[i]), "  insert  affected:", addRows)
		} else if slurmToRtrFullLogs[i].Style == "prefixFilters" {
			labRpkiRtrFullLog := new(model.LabRpkiRtrFullLog)
			engEffect := xormdb.XormEngine.Table(tableName).Where(" serialNumber= ? ", newSerialNumber)
			defer engEffect.Close()
			session = session.Table(tableName).Where(" serialNumber= ? ", newSerialNumber)
			if slurmToRtrFullLogs[i].Asn.Valid {
				engEffect = engEffect.And(` asn = ? `, slurmToRtrFullLogs[i].Asn.ValueOrZero())
				session = session.And(` asn = ? `, slurmToRtrFullLogs[i].Asn.ValueOrZero())
			}
			if slurmToRtrFullLogs[i].PrefixLength > 0 {
				engEffect = engEffect.And(` prefixLength = ? `, slurmToRtrFullLogs[i].PrefixLength)
				session = session.And(` prefixLength = ? `, slurmToRtrFullLogs[i].PrefixLength)
			}
			if slurmToRtrFullLogs[i].MaxLength > 0 {
				engEffect = engEffect.And(` maxLength = ? `, slurmToRtrFullLogs[i].MaxLength)
				session = session.And(` maxLength = ? `, slurmToRtrFullLogs[i].MaxLength)
			}
			if len(slurmToRtrFullLogs[i].Address) > 0 {
				address, _ := iputil.TrimAddressPrefixZero(slurmToRtrFullLogs[i].Address, iputil.GetIpType(slurmToRtrFullLogs[i].Address))
				engEffect = engEffect.And(` address = ? `, address)
				session = session.And(` address = ? `, address)
			}
			if getEffectSlurm {
				effectSlurmToRtrFullLogsTmp := make([]model.SlurmToRtrFullLog, 0)
				err = engEffect.Cols("id,asn,address,prefixLength,maxLength,sourceFrom as sourceFromJson").Find(&effectSlurmToRtrFullLogsTmp)
				if err != nil {
					belogs.Error("UpdateRtrFullOrFullLogFromSlurmDb(): get effectSlurmToRtrFullLogsTmp fail:", err)
					return nil, err
				}
				belogs.Debug("UpdateRtrFullOrFullLogFromSlurmDb(): len(effectSlurmToRtrFullLogsTmp):", len(effectSlurmToRtrFullLogsTmp),
					"   effectSlurmToRtrFullLogsTmp:", jsonutil.MarshalJson(effectSlurmToRtrFullLogsTmp))
				effectSlurmToRtrFullLogs = append(effectSlurmToRtrFullLogs, effectSlurmToRtrFullLogsTmp...)
			}
			affected, err := session.Delete(labRpkiRtrFullLog)
			if err != nil {
				belogs.Error("UpdateRtrFullOrFullLogFromSlurmDb():del lab_rpki_rtr_full_log/lab_rpki_rtr_full from slurm fail,  slurmToRtrFullLogs:",
					jsonutil.MarshalJson(slurmToRtrFullLogs[i]), affected, err)
				return nil, xormdb.RollbackAndLogError(session, "UpdateRtrFullOrFullLogFromSlurmDb(): :del lab_rpki_rtr_full_log/lab_rpki_rtr_full from slurm fail: ", err)
			}

			belogs.Debug("UpdateRtrFullOrFullLogFromSlurmDb(): delete lab_rpki_rtr_full_log/lab_rpki_rtr_full from slurmToRtrFullLogs, newSerialNumber:",
				newSerialNumber, ",  slurmToRtrFullLogs[i]: ", jsonutil.MarshalJson(slurmToRtrFullLogs[i]), " delete  affected:", affected)
		}
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateRtrFullOrFullLogFromSlurmDb(): CommitSession fail :", err)
		return nil, xormdb.RollbackAndLogError(session, "UpdateRtrFullOrFullLogFromSlurmDb(): CommitSession fail: ", err)
	}
	belogs.Info("UpdateRtrFullOrFullLogFromSlurmDb():CommitSession ok,  len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs),
		"   len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), " time(s):", time.Since(start))
	belogs.Debug("UpdateRtrFullOrFullLogFromSlurmDb():effectSlurmToRtrFullLogs:", jsonutil.MarshalJson(effectSlurmToRtrFullLogs))
	return effectSlurmToRtrFullLogs, nil
}

// tableName:lab_rpki_rtr_asa_full_log / lab_rpki_rtr_asa_full
func UpdateRtrAsaFullOrFullLogFromSlurmDb(tableName string, newSerialNumber uint64,
	slurmToRtrFullLogs []model.SlurmToRtrFullLog, getEffectSlurm bool) (effectSlurmToRtrFullLogs []model.SlurmToRtrFullLog, err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()
	effectSlurmToRtrFullLogs = make([]model.SlurmToRtrFullLog, 0)

	// insert ignore into rtr_asa_full_log
	sqlInsertSlurm := `insert   into ` + tableName + `
				(serialNumber,customerAsn,providerAsn,addressFamily,sourceFrom) values
				(?,?,?,  ?,?)`

	belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): will insert/del lab_rpki_rtr_asa_full_log/lab_rpki_rtr_asa_full from slurmToRtrFullLogs,sqlInsertSlurm:", sqlInsertSlurm,
		" newSerialNumber:", newSerialNumber, "    len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))
	for i := range slurmToRtrFullLogs {
		sourceFrom := model.LabRpkiRtrSourceFrom{
			Source:         "slurm",
			SlurmId:        slurmToRtrFullLogs[i].SlurmId,
			SlurmLogId:     slurmToRtrFullLogs[i].SlurmLogId,
			SlurmLogFileId: slurmToRtrFullLogs[i].SlurmLogFileId,
		}
		sourceFormJson := jsonutil.MarshalJson(sourceFrom)
		slurmToRtrFullLogs[i].SourceFromJson = sourceFormJson
		belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): slurmToRtrFullLogs[i]:", jsonutil.MarshalJson(slurmToRtrFullLogs[i]),
			" sourceFrom:", sourceFrom, "  i:", i)
		customerAsn := slurmToRtrFullLogs[i].CustomerAsn
		providerAsn := slurmToRtrFullLogs[i].ProviderAsn
		addressFamilyIpv4, addressFamilyIpv6, err := ConvertSlurmAddressFamilyToRtr(slurmToRtrFullLogs[i].AddressFamily)
		if err != nil {
			belogs.Error("UpdateRtrAsaFullOrFullLogFromSlurmDb(): ConvertSlurmAddressFamilyToRtr fail :", err)
			return nil, xormdb.RollbackAndLogError(session, "UpdateRtrAsaFullOrFullLogFromSlurmDb(): ConvertSlurmAddressFamilyToRtr fail: ", err)
		}

		if slurmToRtrFullLogs[i].Style == "aspaAssertions" {

			belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions from slurm, i:", i, "  slurmToRtrFullLogs:",
				"  newSerialNumber:", newSerialNumber, "  customerAsn: ", customerAsn, "  providerAsn:", providerAsn,
				"  addressFamilyIpv4:", addressFamilyIpv4, "  addressFamilyIpv6:", addressFamilyIpv6,
				"  sourceFormJson:", sourceFormJson)

			if addressFamilyIpv4.Valid {
				affected, err := session.Exec(sqlInsertSlurm,
					newSerialNumber,
					customerAsn,
					providerAsn,
					addressFamilyIpv4,
					sourceFormJson)

				if err != nil {
					belogs.Error("UpdateRtrAsaFullOrFullLogFromSlurmDb():aspaAssertions from slurm ipv4 fail:",
						"  newSerialNumber:", newSerialNumber, "  customerAsn:", customerAsn, "  providerAsn:", providerAsn,
						"  addressFamilyIpv4:", addressFamilyIpv4, "  sourceFormJson:", sourceFormJson,
						"  affected:", affected, err)
					return nil, xormdb.RollbackAndLogError(session, "UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions insert into lab_rpki_rtr_asa_full_log/lab_rpki_rtr_asa_full from slurm ipv4 fail: ", err)
				}
				addRows, _ := affected.RowsAffected()
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions from slurm ipv4, slurmToRtrFullLogs:",
					"  newSerialNumber:", newSerialNumber, "  customerAsn: ", customerAsn, "  providerAsn:", providerAsn,
					"  addressFamilyIpv4:", addressFamilyIpv4, "  sourceFormJson:", sourceFormJson,
					"  insert  affected:", addRows)

			}
			if addressFamilyIpv6.Valid {
				affected, err := session.Exec(sqlInsertSlurm,
					newSerialNumber,
					customerAsn,
					providerAsn,
					addressFamilyIpv6,
					sourceFormJson)

				if err != nil {
					belogs.Error("UpdateRtrAsaFullOrFullLogFromSlurmDb():aspaAssertions insert into lab_rpki_rtr_asa_full_log/lab_rpki_rtr_asa_full from slurm ipv6 fail:",
						"  newSerialNumber:", newSerialNumber, "  customerAsn:", customerAsn, "  providerAsn:", providerAsn,
						"  addressFamilyIpv6:", addressFamilyIpv6, "  sourceFormJson:", sourceFormJson,
						"  affected:", affected, err)
					return nil, xormdb.RollbackAndLogError(session, "UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions insert into lab_rpki_rtr_asa_full_log/lab_rpki_rtr_asa_full from slurm ipv6 fail: ", err)
				}
				addRows, _ := affected.RowsAffected()
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions from slurm ipv6, slurmToRtrFullLogs:",
					"  newSerialNumber:", newSerialNumber, "  customerAsn: ", customerAsn, "  providerAsn:", providerAsn,
					"  addressFamilyIpv4:", addressFamilyIpv4, "  sourceFormJson:", sourceFormJson,
					"  insert  affected:", addRows)
			}

			if getEffectSlurm {
				effectSlurmToRtrFullLogs = append(effectSlurmToRtrFullLogs, slurmToRtrFullLogs[i])
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions getEffectSlurm, slurmToRtrFullLogs:", jsonutil.MarshalJson(slurmToRtrFullLogs[i]))
			}

		} else if slurmToRtrFullLogs[i].Style == "aspaFilters" {
			belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, i:", i, "  slurmToRtrFullLogs:",
				"  newSerialNumber:", newSerialNumber, "  customerAsn: ", customerAsn, "  providerAsn:", providerAsn,
				"  addressFamilyIpv4:", addressFamilyIpv4, "  addressFamilyIpv6:", addressFamilyIpv6,
				"  sourceFormJson:", sourceFormJson)

			labRpkiRtrAsaFullLog := new(model.LabRpkiRtrAsaFullLog)
			engEffect := xormdb.XormEngine.Table(tableName).Where(" serialNumber= ? ", newSerialNumber)
			session = session.Table(tableName).Where(" serialNumber= ? ", newSerialNumber)

			if customerAsn.Valid {
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, customerAsn.Valid, i:", i, "  customerAsn:", customerAsn)
				engEffect = engEffect.And(` customerAsn = ? `, customerAsn)
				session = session.And(` customerAsn = ? `, customerAsn)
			}
			if providerAsn.Valid {
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, providerAsn.Valid, i:", i, "  providerAsn:", providerAsn)
				engEffect = engEffect.And(` providerAsn = ? `, providerAsn.ValueOrZero())
				session = session.And(` providerAsn = ? `, providerAsn.ValueOrZero())
			}

			if addressFamilyIpv4.Valid && addressFamilyIpv6.Valid {
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, Ipv4 and Ipv6.Valid, i:", i,
					"  addressFamilyIpv4:", addressFamilyIpv4, "   addressFamilyIpv6:", addressFamilyIpv6)
				engEffect = engEffect.In(`addressFamily`, addressFamilyIpv4.ValueOrZero(), addressFamilyIpv6.ValueOrZero())
				session = session.In(`addressFamily`, addressFamilyIpv4.ValueOrZero(), addressFamilyIpv6.ValueOrZero())
			} else if addressFamilyIpv4.Valid && !addressFamilyIpv6.Valid {
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, Ipv4 Valid, i:", i,
					"  addressFamilyIpv4:", addressFamilyIpv4, "   addressFamilyIpv6:", addressFamilyIpv6)
				engEffect = engEffect.In(`addressFamily`, addressFamilyIpv4.ValueOrZero())
				session = session.In(`addressFamily`, addressFamilyIpv4.ValueOrZero())
			} else if !addressFamilyIpv4.Valid && addressFamilyIpv6.Valid {
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, Ipv6 Valid, i:", i,
					"  addressFamilyIpv4:", addressFamilyIpv4, "   addressFamilyIpv6:", addressFamilyIpv6)
				engEffect = engEffect.In(`addressFamily`, addressFamilyIpv6.ValueOrZero())
				session = session.In(`addressFamily`, addressFamilyIpv6.ValueOrZero())
			}

			if getEffectSlurm {
				effectSlurmToRtrFullLogsTmp := make([]model.SlurmToRtrFullLog, 0)
				err = engEffect.Cols("id,customerAsn,providerAsn,addressFamily,sourceFrom as sourceFromJson").Find(&effectSlurmToRtrFullLogsTmp)
				if err != nil {
					belogs.Error("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters effectSlurmToRtrFullLogsTmp fail:", err)
					return nil, err
				}
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters len(effectSlurmToRtrFullLogsTmp):", len(effectSlurmToRtrFullLogsTmp),
					"   effectSlurmToRtrFullLogsTmp:", jsonutil.MarshalJson(effectSlurmToRtrFullLogsTmp))
				effectSlurmToRtrFullLogs = append(effectSlurmToRtrFullLogs, effectSlurmToRtrFullLogsTmp...)

			}

			affected, err := session.Delete(labRpkiRtrAsaFullLog)
			if err != nil {
				belogs.Error("UpdateRtrAsaFullOrFullLogFromSlurmDb():aspaFilters del from slurm fail,  slurmToRtrFullLogs:",
					jsonutil.MarshalJson(slurmToRtrFullLogs[i]), affected, err)
				return nil, xormdb.RollbackAndLogError(session, "UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters del from slurm fail: ", err)
			}

			belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters delete from slurmToRtrFullLogs ",
				"  newSerialNumber:", newSerialNumber, "  customerAsn: ", customerAsn, "  providerAsn:", providerAsn,
				"  addressFamilyIpv4:", addressFamilyIpv4, "  addressFamilyIpv6:", addressFamilyIpv6,
				"  sourceFormJson:", sourceFormJson, " delete  affected:", affected)
		}
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateRtrAsaFullOrFullLogFromSlurmDb(): CommitSession fail :", err)
		return nil, xormdb.RollbackAndLogError(session, "UpdateRtrAsaFullOrFullLogFromSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("UpdateRtrAsaFullOrFullLogFromSlurmDb():CommitSession ok,  len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs),
		"   len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), " time(s):", time.Since(start))
	belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb():effectSlurmToRtrFullLogs:", jsonutil.MarshalJson(effectSlurmToRtrFullLogs))
	return effectSlurmToRtrFullLogs, nil
}

/*
// tableName:rtr_asa_full_log / rtr_asa_full
func UpdateRtrAsaFullOrFullLogFromSlurmDb(tableName string, newSerialNumber uint64,
	slurmToRtrFullLogs []model.SlurmToRtrFullLog, getEffectSlurm bool) (effectSlurmToRtrFullLogs []model.SlurmToRtrFullLog, err error) {
	start := time.Now()
	session, err := xormdb.NewSession()
	defer session.Close()
	effectSlurmToRtrFullLogs = make([]model.SlurmToRtrFullLog, 0)

	// insert ignore into rtr_asa_full_log
	sqlInsertSlurm := `insert   into ` + tableName + `
				(serialNumber,customerAsn,providerAsn,addressFamily,sourceFrom) values
				(?,?,?,  ?,?)`

	belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): will insert/del asa_full_log/asa_full from slurmToRtrFullLogs,sqlInsertSlurm:", sqlInsertSlurm,
		" newSerialNumber:", newSerialNumber, "    len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))
	for i := range slurmToRtrFullLogs {
		sourceFrom := model.LabRpkiRtrSourceFrom{
			Source:         "slurm",
			SlurmId:        slurmToRtrFullLogs[i].SlurmId,
			SlurmLogId:     slurmToRtrFullLogs[i].SlurmLogId,
			SlurmLogFileId: slurmToRtrFullLogs[i].SlurmLogFileId,
		}
		sourceFormJson := jsonutil.MarshalJson(sourceFrom)
		slurmToRtrFullLogs[i].SourceFromJson = sourceFormJson
		belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): slurmToRtrFullLogs[i]:", jsonutil.MarshalJson(slurmToRtrFullLogs[i]),
			" sourceFrom:", sourceFrom, "  i:", i)
		customerAsn := slurmToRtrFullLogs[i].CustomerAsn
		providerAsn := slurmToRtrFullLogs[i].ProviderAsn
		addressFamilyIpv4, addressFamilyIpv6, err := ConvertSlurmAddressFamilyToRtr(slurmToRtrFullLogs[i].AddressFamily)
		if err != nil {
			belogs.Error("UpdateRtrAsaFullOrFullLogFromSlurmDb(): ConvertSlurmAddressFamilyToRtr fail :", err)
			return nil, xormdb.RollbackAndLogError(session, "UpdateRtrAsaFullOrFullLogFromSlurmDb(): ConvertSlurmAddressFamilyToRtr fail: ", err)
		}

		if slurmToRtrFullLogs[i].Style == "aspaAssertions" {

			belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions from slurm, i:", i, "  slurmToRtrFullLogs:",
				"  newSerialNumber:", newSerialNumber, "  customerAsn: ", customerAsn, "  providerAsn:", providerAsn,
				"  addressFamilyIpv4:", addressFamilyIpv4, "  addressFamilyIpv6:", addressFamilyIpv6,
				"  sourceFormJson:", sourceFormJson)

			if addressFamilyIpv4.Valid {
				affected, err := session.Exec(sqlInsertSlurm,
					newSerialNumber,
					customerAsn,
					providerAsn,
					addressFamilyIpv4,
					sourceFormJson)

				if err != nil {
					belogs.Error("UpdateRtrAsaFullOrFullLogFromSlurmDb():aspaAssertions from slurm ipv4 fail:",
						"  newSerialNumber:", newSerialNumber, "  customerAsn:", customerAsn, "  providerAsn:", providerAsn,
						"  addressFamilyIpv4:", addressFamilyIpv4, "  sourceFormJson:", sourceFormJson,
						"  affected:", affected, err)
					return nil, xormdb.RollbackAndLogError(session, "UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions insert into lab_rpki_rtr_asa_full_log/lab_rpki_rtr_asa_full from slurm ipv4 fail: ", err)
				}
				addRows, _ := affected.RowsAffected()
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions from slurm ipv4, slurmToRtrFullLogs:",
					"  newSerialNumber:", newSerialNumber, "  customerAsn: ", customerAsn, "  providerAsn:", providerAsn,
					"  addressFamilyIpv4:", addressFamilyIpv4, "  sourceFormJson:", sourceFormJson,
					"  insert  affected:", addRows)

			}
			if addressFamilyIpv6.Valid {
				affected, err := session.Exec(sqlInsertSlurm,
					newSerialNumber,
					customerAsn,
					providerAsn,
					addressFamilyIpv6,
					sourceFormJson)

				if err != nil {
					belogs.Error("UpdateRtrAsaFullOrFullLogFromSlurmDb():aspaAssertions insert into lab_rpki_rtr_asa_full_log/lab_rpki_rtr_asa_full from slurm ipv6 fail:",
						"  newSerialNumber:", newSerialNumber, "  customerAsn:", customerAsn, "  providerAsn:", providerAsn,
						"  addressFamilyIpv6:", addressFamilyIpv6, "  sourceFormJson:", sourceFormJson,
						"  affected:", affected, err)
					return nil, xormdb.RollbackAndLogError(session, "UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions insert into lab_rpki_rtr_asa_full_log/lab_rpki_rtr_asa_full from slurm ipv6 fail: ", err)
				}
				addRows, _ := affected.RowsAffected()
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions from slurm ipv6, slurmToRtrFullLogs:",
					"  newSerialNumber:", newSerialNumber, "  customerAsn: ", customerAsn, "  providerAsn:", providerAsn,
					"  addressFamilyIpv4:", addressFamilyIpv4, "  sourceFormJson:", sourceFormJson,
					"  insert  affected:", addRows)
			}

			if getEffectSlurm {
				effectSlurmToRtrFullLogs = append(effectSlurmToRtrFullLogs, slurmToRtrFullLogs[i])
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaAssertions getEffectSlurm, slurmToRtrFullLogs:", jsonutil.MarshalJson(slurmToRtrFullLogs[i]))
			}

		} else if slurmToRtrFullLogs[i].Style == "aspaFilters" {
			belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, i:", i, "  slurmToRtrFullLogs:",
				"  newSerialNumber:", newSerialNumber, "  customerAsn: ", customerAsn, "  providerAsn:", providerAsn,
				"  addressFamilyIpv4:", addressFamilyIpv4, "  addressFamilyIpv6:", addressFamilyIpv6,
				"  sourceFormJson:", sourceFormJson)

			labRpkiRtrAsaFullLog := new(model.LabRpkiRtrAsaFullLog)
			session = session.Table(tableName).Where(" serialNumber= ? ", newSerialNumber)

			if customerAsn.Valid {
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, customerAsn.Valid, i:", i, "  customerAsn:", customerAsn)
				session = session.And(` customerAsn = ? `, customerAsn)
			}
			if providerAsn.Valid {
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, providerAsn.Valid, i:", i, "  providerAsn:", providerAsn)
				session = session.And(` providerAsn = ? `, providerAsn.ValueOrZero())
			}

			if addressFamilyIpv4.Valid && addressFamilyIpv6.Valid {
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, Ipv4 and Ipv6.Valid, i:", i,
					"  addressFamilyIpv4:", addressFamilyIpv4, "   addressFamilyIpv6:", addressFamilyIpv6)
				session = session.In(`addressFamily`, addressFamilyIpv4.ValueOrZero(), addressFamilyIpv6.ValueOrZero())
			} else if addressFamilyIpv4.Valid && !addressFamilyIpv6.Valid {
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, Ipv4 Valid, i:", i,
					"  addressFamilyIpv4:", addressFamilyIpv4, "   addressFamilyIpv6:", addressFamilyIpv6)
				session = session.In(`addressFamily`, addressFamilyIpv4.ValueOrZero())
			} else if !addressFamilyIpv4.Valid && addressFamilyIpv6.Valid {
				belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters from slurm, Ipv6 Valid, i:", i,
					"  addressFamilyIpv4:", addressFamilyIpv4, "   addressFamilyIpv6:", addressFamilyIpv6)
				session = session.In(`addressFamily`, addressFamilyIpv6.ValueOrZero())
			}

			if getEffectSlurm {

			}

			affected, err := session.Delete(labRpkiRtrAsaFullLog)
			if err != nil {
				belogs.Error("UpdateRtrAsaFullOrFullLogFromSlurmDb():aspaFilters del from slurm fail,  slurmToRtrFullLogs:",
					jsonutil.MarshalJson(slurmToRtrFullLogs[i]), affected, err)
				return nil, xormdb.RollbackAndLogError(session, "UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters del from slurm fail: ", err)
			}

			belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb(): aspaFilters delete from slurmToRtrFullLogs ",
				"  newSerialNumber:", newSerialNumber, "  customerAsn: ", customerAsn, "  providerAsn:", providerAsn,
				"  addressFamilyIpv4:", addressFamilyIpv4, "  addressFamilyIpv6:", addressFamilyIpv6,
				"  sourceFormJson:", sourceFormJson, " delete  affected:", affected)
		}
	}
	// commit
	err = xormdb.CommitSession(session)
	if err != nil {
		belogs.Error("UpdateRtrAsaFullOrFullLogFromSlurmDb(): CommitSession fail :", err)
		return nil, xormdb.RollbackAndLogError(session, "UpdateRtrAsaFullOrFullLogFromSlurmDb(): CommitSession fail: ", err)
	}

	belogs.Info("UpdateRtrAsaFullOrFullLogFromSlurmDb():CommitSession ok,  len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs),
		"   len(effectSlurmToRtrFullLogs):", len(effectSlurmToRtrFullLogs), " time(s):", time.Since(start))
	belogs.Debug("UpdateRtrAsaFullOrFullLogFromSlurmDb():effectSlurmToRtrFullLogs:", jsonutil.MarshalJson(effectSlurmToRtrFullLogs))
	return effectSlurmToRtrFullLogs, nil
}
*/
// style=prefix/asa
func GetAllSlurmsDb(style string) (slurmToRtrFullLogs []model.SlurmToRtrFullLog, err error) {
	// get all slurm, not care state->"$.rtr"='notYet' or 'finished'
	var sql string
	if style == "prefix" {
		sql = `select id as slurmId, style, 
		    asn,
			substring_index( addressPrefix, '/', 1 ) AS address,
			substring_index( addressPrefix, '/', -1 ) AS prefixLength,
			maxLength, 
			slurmLogId,
		    slurmLogFileId 
	    from lab_rpki_slurm  where style in ('prefixFilters','prefixAssertions')
		order by id `
	} else if style == "asa" {
		sql = `select id as slurmId, style,  
			customerAsn,
			providerAsn, 
			addressFamily,
			slurmLogId,
		    slurmLogFileId 
	    from lab_rpki_slurm  where style in ('aspaFilters','aspaAssertions') 
		order by id `
	}
	belogs.Debug("GetAllSlurmsDb(): sql:", sql)
	err = xormdb.XormEngine.SQL(sql).Find(&slurmToRtrFullLogs)
	if err != nil {
		belogs.Error("GetAllSlurmsDb(): find fail:", err)
		return slurmToRtrFullLogs, err
	}
	belogs.Debug("GetAllSlurmsDb(): len(slurmToRtrFullLogs):", len(slurmToRtrFullLogs))

	return slurmToRtrFullLogs, nil
}

func InsertSerialNumberDb(session *xorm.Session, newSerialNumberModel *SerialNumberModel, start time.Time) error {
	belogs.Debug("InsertSerialNumberDb():newSerialNumberModel.HaveSaveToDb:", atomic.LoadUint32(&newSerialNumberModel.HaveSaveToDb))
	swapped := atomic.CompareAndSwapUint32(&newSerialNumberModel.HaveSaveToDb, 0, 1)
	belogs.Debug("InsertSerialNumberDb():after CAS, newSerialNumberModel.HaveSaveToDb:", atomic.LoadUint32(&newSerialNumberModel.HaveSaveToDb),
		" swapped:", swapped)
	if swapped {
		sql := ` insert ignore into lab_rpki_rtr_serial_number(
		serialNumber,globalSerialNumber,subpartSerialNumber, createTime)
		 values(?,?,?,?)`
		_, err := session.Exec(sql,
			newSerialNumberModel.SerialNumber, newSerialNumberModel.GlobalSerialNumber,
			newSerialNumberModel.SubpartSerialNumber, start)
		if err != nil {
			belogs.Error("InsertSerialNumberDb():insert into lab_rpki_rtr_serial_number fail:", jsonutil.MarshalJson(newSerialNumberModel), err)
			return err
		}
		belogs.Debug("InsertSerialNumberDb():insert into lab_rpki_rtr_serial_number:", jsonutil.MarshalJson(newSerialNumberModel), "  time(s):", time.Since(start))
	} else {
		belogs.Debug("InsertSerialNumberDb():newSerialNumberModel.HaveSaveToDb, swapped, no insert:", swapped)
	}
	return nil
}
