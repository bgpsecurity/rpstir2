package sync

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/stringutil"
	"xorm.io/xorm"
)

func delAsaDb(session *xorm.Session, filePathPrefix string) (err error) {
	start := time.Now()
	belogs.Debug("delAsaDb():will delete lab_rpki_asa by filePathPrefix:", filePathPrefix)

	// get asaIds
	asaIds := make([]int64, 0)
	err = session.SQL("select id from lab_rpki_asa Where filePath like ? ",
		filePathPrefix+"%").Find(&asaIds)
	if err != nil {
		belogs.Error("delAsaDb(): get asaIds fail, filePathPrefix:", filePathPrefix, err)
		return err
	}
	if len(asaIds) == 0 {
		belogs.Debug("delAsaDb(): len(asaIds)==0, filePathPrefix:", filePathPrefix)
		return nil
	}
	asaIdsStr := stringutil.Int64sToInString(asaIds)
	belogs.Debug("delAsaDb():will delete lab_rpki_asa len(asaIds):", len(asaIds), "   filePathPrefix:", filePathPrefix)

	// get customerProviderAsnIds
	customerProviderAsnIds, err := getIdsByParamIdsDb("lab_rpki_asa_customer_provider_asn", "asaId", asaIdsStr)
	if err != nil {
		belogs.Error("delAsaDb(): get customerProviderAsnIds fail, filePathPrefix:", filePathPrefix,
			"   asaIdsStr:", asaIdsStr, err)
		return err
	}
	belogs.Debug("delAsaDb(): len(customerProviderAsnIds):", len(customerProviderAsnIds), "   filePathPrefix:", filePathPrefix)

	// get siaIds
	siaIds, err := getIdsByParamIdsDb("lab_rpki_asa_sia", "asaId", asaIdsStr)
	if err != nil {
		belogs.Error("delAsaDb(): get siaIds fail, filePathPrefix:", filePathPrefix,
			"   asaIdsStr:", asaIdsStr, err)
		return err
	}
	belogs.Debug("delAsaDb(): len(siaIds):", len(siaIds), "   filePathPrefix:", filePathPrefix)

	// get aiaIds
	aiaIds, err := getIdsByParamIdsDb("lab_rpki_asa_aia", "asaId", asaIdsStr)
	if err != nil {
		belogs.Error("delAsaDb(): get aiaIds fail, filePathPrefix:", filePathPrefix,
			"   asaIdsStr:", asaIdsStr, err)
		return err
	}
	belogs.Debug("delAsaDb(): len(aiaIds):", len(aiaIds), "   filePathPrefix:", filePathPrefix)

	// del customerProviderAsnIds
	customerProviderAsnIdsStr := stringutil.Int64sToInString(customerProviderAsnIds)
	if len(customerProviderAsnIdsStr) > 0 {
		belogs.Debug("delAsaDb(): delete lab_rpki_asa_customer_provider_asn, providerAsnIdsStr:", customerProviderAsnIdsStr)
		_, err := session.Exec("delete from lab_rpki_asa_customer_provider_asn  where id in " + customerProviderAsnIdsStr)
		if err != nil {
			belogs.Error("delAsaDb():delete  from lab_rpki_asa_customer_provider_asn fail: customerProviderAsnIdsStr:", customerProviderAsnIdsStr,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del siaIds
	siaIdsStr := stringutil.Int64sToInString(siaIds)
	if len(siaIdsStr) > 0 {
		belogs.Debug("delAsaDb(): delete lab_rpki_asa_sia, siaIdsStr:", siaIdsStr)
		_, err = session.Exec("delete from  lab_rpki_asa_sia  where id in " + siaIdsStr)
		if err != nil {
			belogs.Error("delAsaDb():delete  from lab_rpki_asa_sia fail: siaIdsStr:", siaIdsStr,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del aiaIds
	aiaIdsStr := stringutil.Int64sToInString(aiaIds)
	if len(aiaIdsStr) > 0 {
		belogs.Debug("delAsaDb(): delete lab_rpki_asa_aia, aiaIdsStr:", aiaIdsStr)
		_, err = session.Exec("delete from  lab_rpki_asa_aia  where id in " + aiaIdsStr)
		if err != nil {
			belogs.Error("delAsaDb():delete  from lab_rpki_asa_aia fail: aiaIdsStr:", aiaIdsStr,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del asaIds
	belogs.Debug("delAsaDb(): delete lab_rpki_asa, asaIdsStr:", asaIdsStr)
	_, err = session.Exec("delete from  lab_rpki_asa  where id in " + asaIdsStr)
	if err != nil {
		belogs.Error("delAsaDb():delete  from lab_rpki_asa fail: asaIdsStr:", asaIdsStr,
			"   filePathPrefix:", filePathPrefix, "   err:", err)
		return err
	}

	belogs.Info("delAsaDb():delete lab_rpki_asa ok, by filePathPrefix:", filePathPrefix,
		"  len(asaIds)", len(asaIds), "     time(s):", time.Since(start))
	return nil
}
