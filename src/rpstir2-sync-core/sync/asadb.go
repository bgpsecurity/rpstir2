package sync

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

func delAsaDb(session *xorm.Session, filePathPrefix string) (err error) {
	start := time.Now()
	belogs.Debug("delAsaDb():will delete lab_rpki_asa_*** by filePathPrefix :", filePathPrefix)

	// get asaIds
	asaIds := make([]int64, 0)
	err = session.SQL("select id from lab_rpki_asa Where filePath like ? ",
		filePathPrefix+"%").Find(&asaIds)
	if err != nil {
		belogs.Error("delAsaDb(): get asaIds fail, filePathPrefix: ", filePathPrefix, err)
		return err
	}
	if len(asaIds) == 0 {
		belogs.Debug("delAsaDb(): len(asaIds)==0, filePathPrefix: ", filePathPrefix)
		return nil
	}
	asaIdsStr := xormdb.Int64sToInString(asaIds)
	belogs.Debug("delAsaDb():will delete lab_rpki_asa len(asaIds):", len(asaIds), asaIdsStr,
		"   filePathPrefix:", filePathPrefix)

	// get providerAsnIds
	providerAsnIds, err := getIdsByParamIdsDb("lab_rpki_asa_provider_asn", "asaId", asaIdsStr)
	if err != nil {
		belogs.Error("delAsaDb(): get providerAsnIds fail, filePathPrefix: ", filePathPrefix,
			"   asaIdsStr:", asaIdsStr, err)
		return err
	}
	belogs.Debug("delAsaDb(): len(providerAsnIds):", len(providerAsnIds), "   filePathPrefix:", filePathPrefix,
		"   asaIdsStr:", asaIdsStr)

	// get customerAsnIds
	customerAsnIds, err := getIdsByParamIdsDb("lab_rpki_asa_customer_asn", "asaId", asaIdsStr)
	if err != nil {
		belogs.Error("delAsaDb(): get customerAsnIds fail, filePathPrefix: ", filePathPrefix,
			"   asaIdsStr:", asaIdsStr, err)
		return err
	}
	belogs.Debug("delAsaDb(): len(customerAsnIds):", len(customerAsnIds), "   filePathPrefix:", filePathPrefix,
		"   asaIdsStr:", asaIdsStr)

	// get siaIds
	siaIds, err := getIdsByParamIdsDb("lab_rpki_asa_sia", "asaId", asaIdsStr)
	if err != nil {
		belogs.Error("delAsaDb(): get siaIds fail, filePathPrefix: ", filePathPrefix,
			"   asaIdsStr:", asaIdsStr, err)
		return err
	}
	belogs.Debug("delAsaDb(): len(siaIds):", len(siaIds), "   filePathPrefix:", filePathPrefix,
		"   asaIdsStr:", asaIdsStr)

	// get aiaIds
	aiaIds, err := getIdsByParamIdsDb("lab_rpki_asa_aia", "asaId", asaIdsStr)
	if err != nil {
		belogs.Error("delAsaDb(): get aiaIds fail, filePathPrefix: ", filePathPrefix,
			"   asaIdsStr:", asaIdsStr, err)
		return err
	}
	belogs.Debug("delAsaDb(): len(aiaIds):", len(aiaIds), "   filePathPrefix:", filePathPrefix,
		"   asaIdsStr:", asaIdsStr)

	// del providerAsnIds
	providerAsnIdsStr := xormdb.Int64sToInString(providerAsnIds)
	if len(providerAsnIdsStr) > 0 {
		_, err := session.Exec("delete from lab_rpki_asa_provider_asn  where id in " + providerAsnIdsStr)
		if err != nil {
			belogs.Error("delAsaDb():delete  from lab_rpki_asa_provider_asn fail: providerAsnIds: ", providerAsnIds,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del customerAsnIds
	customerAsnIdsStr := xormdb.Int64sToInString(customerAsnIds)
	if len(providerAsnIdsStr) > 0 {
		_, err = session.Exec("delete from lab_rpki_asa_customer_asn  where id in " + customerAsnIdsStr)
		if err != nil {
			belogs.Error("delAsaDb():delete  from lab_rpki_asa_customer_asn fail: customerAsnIds: ", customerAsnIds,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del siaIds
	siaIdsStr := xormdb.Int64sToInString(siaIds)
	if len(providerAsnIdsStr) > 0 {
		_, err = session.Exec("delete from  lab_rpki_asa_sia  where id in " + siaIdsStr)
		if err != nil {
			belogs.Error("delAsaDb():delete  from lab_rpki_asa_sia fail: siaIdsStr: ", siaIdsStr,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del aiaIds
	aiaIdsStr := xormdb.Int64sToInString(aiaIds)
	if len(aiaIdsStr) > 0 {
		_, err = session.Exec("delete from  lab_rpki_asa_aia  where id in " + aiaIdsStr)
		if err != nil {
			belogs.Error("delAsaDb():delete  from lab_rpki_asa_aia fail: aiaIdsStr: ", aiaIdsStr,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del asaIds
	_, err = session.Exec("delete from  lab_rpki_asa  where id in " + asaIdsStr)
	if err != nil {
		belogs.Error("delAsaDb():delete  from lab_rpki_asa fail: asaIdsStr: ", asaIdsStr,
			"   filePathPrefix:", filePathPrefix, "   err:", err)
		return err
	}

	belogs.Info("delAsaDb():delete lab_rpki_asa_*** ok, by filePathPrefix :", filePathPrefix,
		"  len(asaIds)", len(asaIds), "     time(s):", time.Since(start))
	return nil
}
