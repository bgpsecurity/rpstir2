package sync

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

func DelRoa(session *xorm.Session, filePathPrefix string) (err error) {
	start := time.Now()
	belogs.Debug("DelRoa():will delete lab_rpki_roa_*** by filePathPrefix :", filePathPrefix)

	// get roaIds
	roaIds := make([]int64, 0)
	err = session.SQL("select id from lab_rpki_roa Where filePath like ? ",
		filePathPrefix+"%").Find(&roaIds)
	if err != nil {
		belogs.Error("DelRoa(): get roaIds fail, filePathPrefix: ", filePathPrefix, err)
		return err
	}
	if len(roaIds) == 0 {
		belogs.Debug("DelRoa(): len(roaIds)==0, filePathPrefix: ", filePathPrefix)
		return nil
	}
	roaIdsStr := xormdb.Int64sToInString(roaIds)
	belogs.Debug("DelRoa():will delete lab_rpki_roa len(roaIds):", len(roaIds), roaIdsStr,
		"   filePathPrefix:", filePathPrefix)

	// get ipIds
	ipIds, err := getIdsByParamIds("lab_rpki_roa_ipaddress", "roaId", roaIdsStr)
	if err != nil {
		belogs.Error("DelRoa(): get ipIds fail, filePathPrefix: ", filePathPrefix,
			"   roaIdsStr:", roaIdsStr, err)
		return err
	}
	belogs.Debug("DelRoa(): len(ipIds):", len(ipIds), "   filePathPrefix:", filePathPrefix,
		"   roaIdsStr:", roaIdsStr)

	// get eeIpIds
	eeIpIds, err := getIdsByParamIds("lab_rpki_roa_ee_ipaddress", "roaId", roaIdsStr)
	if err != nil {
		belogs.Error("DelRoa(): get eeIpIds fail, filePathPrefix: ", filePathPrefix,
			"   roaIdsStr:", roaIdsStr, err)
		return err
	}
	belogs.Debug("DelRoa(): len(eeIpIds):", len(eeIpIds), "   filePathPrefix:", filePathPrefix,
		"   roaIdsStr:", roaIdsStr)

	// get siaIds
	siaIds, err := getIdsByParamIds("lab_rpki_roa_sia", "roaId", roaIdsStr)
	if err != nil {
		belogs.Error("DelRoa(): get siaIds fail, filePathPrefix: ", filePathPrefix,
			"   roaIdsStr:", roaIdsStr, err)
		return err
	}
	belogs.Debug("DelRoa(): len(siaIds):", len(siaIds), "   filePathPrefix:", filePathPrefix,
		"   roaIdsStr:", roaIdsStr)

	// get aiaIds
	aiaIds, err := getIdsByParamIds("lab_rpki_roa_aia", "roaId", roaIdsStr)
	if err != nil {
		belogs.Error("DelRoa(): get aiaIds fail, filePathPrefix: ", filePathPrefix,
			"   roaIdsStr:", roaIdsStr, err)
		return err
	}
	belogs.Debug("DelRoa(): len(aiaIds):", len(aiaIds), "   filePathPrefix:", filePathPrefix,
		"   roaIdsStr:", roaIdsStr)

	// del ipIds
	ipIdsStr := xormdb.Int64sToInString(ipIds)
	if len(ipIdsStr) > 0 {
		_, err := session.Exec("delete from lab_rpki_roa_ipaddress  where id in " + ipIdsStr)
		if err != nil {
			belogs.Error("DelRoa():delete  from lab_rpki_roa_ipaddress fail: ipIds: ", ipIds,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del eeIpIds
	eeIpIdsStr := xormdb.Int64sToInString(eeIpIds)
	if len(ipIdsStr) > 0 {
		_, err = session.Exec("delete from lab_rpki_roa_ee_ipaddress  where id in " + eeIpIdsStr)
		if err != nil {
			belogs.Error("DelRoa():delete  from lab_rpki_roa_ee_ipaddress fail: eeIpIds: ", eeIpIds,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del siaIds
	siaIdsStr := xormdb.Int64sToInString(siaIds)
	if len(ipIdsStr) > 0 {
		_, err = session.Exec("delete from  lab_rpki_roa_sia  where id in " + siaIdsStr)
		if err != nil {
			belogs.Error("DelRoa():delete  from lab_rpki_roa_sia fail: siaIdsStr: ", siaIdsStr,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del aiaIds
	aiaIdsStr := xormdb.Int64sToInString(aiaIds)
	if len(aiaIdsStr) > 0 {
		_, err = session.Exec("delete from  lab_rpki_roa_aia  where id in " + aiaIdsStr)
		if err != nil {
			belogs.Error("DelRoa():delete  from lab_rpki_roa_aia fail: aiaIdsStr: ", aiaIdsStr,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del roaIds
	_, err = session.Exec("delete from  lab_rpki_roa  where id in " + roaIdsStr)
	if err != nil {
		belogs.Error("DelRoa():delete  from lab_rpki_roa fail: roaIdsStr: ", roaIdsStr,
			"   filePathPrefix:", filePathPrefix, "   err:", err)
		return err
	}

	belogs.Debug("DelRoa():delete lab_rpki_roa_*** ok, by filePathPrefix :", filePathPrefix,
		"  len(roaIds)", len(roaIds), "     time(s):", time.Now().Sub(start).Seconds())
	return nil
}
