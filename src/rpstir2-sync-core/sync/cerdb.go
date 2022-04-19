package sync

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

func delCerDb(session *xorm.Session, filePathPrefix string) (err error) {
	start := time.Now()
	belogs.Debug("delCerDb():will delete lab_rpki_cer_*** by filePathPrefix :", filePathPrefix)

	// get cer from lab_rpki_cer
	cerIds := make([]int64, 0)
	err = xormdb.XormEngine.SQL("select id from lab_rpki_cer Where filePath like ? ",
		filePathPrefix+"%").Find(&cerIds)
	if err != nil {
		belogs.Error("delCerDb(): get cerIds fail, filePathPrefix: ", filePathPrefix, err)
		return err
	}
	if len(cerIds) == 0 {
		belogs.Debug("delCerDb(): len(cerIds)==0, filePathPrefix: ", filePathPrefix)
		return nil
	}
	cerIdsStr := xormdb.Int64sToInString(cerIds)
	belogs.Debug("delCerDb():will delete lab_rpki_cer len(cerIds):", len(cerIds), cerIdsStr,
		"   filePathPrefix:", filePathPrefix)

	// get siaIds
	siaIds, err := getIdsByParamIdsDb("lab_rpki_cer_sia", "cerId", cerIdsStr)
	if err != nil {
		belogs.Error("delCerDb(): get siaIds fail, filePathPrefix: ", filePathPrefix,
			"   cerIdsStr:", cerIdsStr, err)
		return err
	}
	belogs.Debug("delCerDb(): len(siaIds):", len(siaIds), "   filePathPrefix:", filePathPrefix,
		"   cerIdsStr:", cerIdsStr)

	// get ipIds
	ipIds, err := getIdsByParamIdsDb("lab_rpki_cer_ipaddress", "cerId", cerIdsStr)
	if err != nil {
		belogs.Error("delCerDb(): get ipIds fail, filePathPrefix: ", filePathPrefix,
			"   cerIdsStr:", cerIdsStr, err)
		return err
	}
	belogs.Debug("delCerDb(): len(ipIds):", len(ipIds), "   filePathPrefix:", filePathPrefix,
		"   cerIdsStr:", cerIdsStr)

	// get crldpIds
	crldpIds, err := getIdsByParamIdsDb("lab_rpki_cer_crldp", "cerId", cerIdsStr)
	if err != nil {
		belogs.Error("delCerDb(): get crldpIds fail, filePathPrefix: ", filePathPrefix,
			"   cerIdsStr:", cerIdsStr, err)
		return err
	}
	belogs.Debug("delCerDb(): len(crldpIds):", len(crldpIds), "   filePathPrefix:", filePathPrefix,
		"   cerIdsStr:", cerIdsStr)

	// get asnIds
	asnIds, err := getIdsByParamIdsDb("lab_rpki_cer_asn", "cerId", cerIdsStr)
	if err != nil {
		belogs.Error("delCerDb(): get asnIds fail, filePathPrefix: ", filePathPrefix,
			"   cerIdsStr:", cerIdsStr, err)
		return err
	}
	belogs.Debug("delCerDb(): len(asnIds):", len(asnIds), "   filePathPrefix:", filePathPrefix,
		"   cerIdsStr:", cerIdsStr)

	// get aiaIds
	aiaIds, err := getIdsByParamIdsDb("lab_rpki_cer_aia", "cerId", cerIdsStr)
	if err != nil {
		belogs.Error("delCerDb(): get aiaIds fail, filePathPrefix: ", filePathPrefix,
			"   cerIdsStr:", cerIdsStr, err)
		return err
	}
	belogs.Debug("delCerDb(): len(aiaIds):", len(aiaIds), "   filePathPrefix:", filePathPrefix,
		"   cerIdsStr:", cerIdsStr)

	// del siaIds
	siaIdsStr := xormdb.Int64sToInString(siaIds)
	if len(siaIdsStr) > 0 {
		_, err = session.Exec("delete from lab_rpki_cer_sia where id in " + siaIdsStr)
		if err != nil {
			belogs.Error("delCerDb():delete from lab_rpki_cer_sia failed, siaIdsStr:", siaIdsStr,
				"   filePathPrefix:", filePathPrefix, "  err:", err)
			return err
		}
	}

	// del ipIds
	ipIdsStr := xormdb.Int64sToInString(ipIds)
	if len(ipIdsStr) > 0 {
		_, err = session.Exec("delete from  lab_rpki_cer_ipaddress  where id in " + ipIdsStr)
		if err != nil {
			belogs.Error("delCerDb():delete  from lab_rpki_cer_ipaddress failed, ipIdsStr:", ipIdsStr,
				"   filePathPrefix:", filePathPrefix, "     err:", err)
			return err
		}
	}

	// del crldpIds
	crldpIdsStr := xormdb.Int64sToInString(crldpIds)
	if len(crldpIdsStr) > 0 {
		_, err = session.Exec("delete  from lab_rpki_cer_crldp  where id in " + crldpIdsStr)
		if err != nil {
			belogs.Error("delCerDb():delete  from lab_rpki_cer_crldp failed, ipIdsStr:", ipIdsStr,
				"   filePathPrefix:", filePathPrefix, "     err:", err)
			return err
		}
	}

	// del asnIds
	asnIdsStr := xormdb.Int64sToInString(asnIds)
	if len(asnIdsStr) > 0 {
		_, err = session.Exec("delete  from lab_rpki_cer_asn  where id in " + asnIdsStr)
		if err != nil {
			belogs.Error("delCerDb():delete  from lab_rpki_cer_asn  failed, asnIdsStr:", asnIdsStr,
				"   filePathPrefix:", filePathPrefix, "     err:", err)
			return err
		}
	}

	// del aiaIds
	aiaIdsStr := xormdb.Int64sToInString(aiaIds)
	if len(aiaIdsStr) > 0 {
		_, err = session.Exec("delete  from lab_rpki_cer_aia  where id in " + aiaIdsStr)
		if err != nil {
			belogs.Error("delCerDb():delete  from lab_rpki_cer_aia  failed, aiaIdsStr:", aiaIdsStr,
				"   filePathPrefix:", filePathPrefix, "     err:", err)
			return err
		}
	}

	// del cer
	_, err = session.Exec("delete  from lab_rpki_cer  where id in " + cerIdsStr)
	if err != nil {
		belogs.Error("delCerDb():delete  from lab_rpki_cer  failed, cerIdsStr:", cerIdsStr,
			"   filePathPrefix:", filePathPrefix, "     err:", err)
		return err
	}
	belogs.Info("delCerDb():delete lab_rpki_cer_*** ok, by filePathPrefix :", filePathPrefix,
		"  len(cerIds)", len(cerIds), "     time(s):", time.Now().Sub(start).Seconds())
	return nil
}
