package sync

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

func delCrlDb(session *xorm.Session, filePathPrefix string) (err error) {
	start := time.Now()
	belogs.Debug("delCrlDb():will delete lab_rpki_crl_*** by filePathPrefix :", filePathPrefix)

	// get crlIds
	crlIds := make([]int64, 0)
	err = session.SQL("select id from lab_rpki_crl Where filePath like ? ",
		filePathPrefix+"%").Find(&crlIds)
	if err != nil {
		belogs.Error("delCrlDb(): get crlIds fail,  filePathPrefix:", filePathPrefix, "     err:", err)
		return err
	}
	if len(crlIds) == 0 {
		belogs.Debug("delCrlDb(): len(crlIds)==0, filePathPrefix: ", filePathPrefix)
		return nil
	}
	crlIdsStr := xormdb.Int64sToInString(crlIds)
	belogs.Debug("delCrlDb():will delete lab_rpki_crl len(crlIds):", len(crlIds), crlIdsStr,
		"   filePathPrefix:", filePathPrefix)

	// get revokeIds
	revokeIds, err := getIdsByParamIdsDb("lab_rpki_crl_revoked_cert", "crlId", crlIdsStr)
	if err != nil {
		belogs.Error("delCrlDb(): get revokeIds fail, filePathPrefix: ", filePathPrefix,
			"   crlIdsStr:", crlIdsStr, err)
		return err
	}
	belogs.Debug("delCrlDb(): len(revokeIds):", len(revokeIds), "   filePathPrefix:", filePathPrefix,
		"   crlIdsStr:", crlIdsStr)

	// del revokeIds
	revokeIdsStr := xormdb.Int64sToInString(revokeIds)
	if len(revokeIdsStr) > 0 {
		_, err := session.Exec("delete from lab_rpki_crl_revoked_cert where id in " + revokeIdsStr)
		if err != nil {
			belogs.Error("delCrlDb():delete  from lab_rpki_crl_revoked_cert fail: revokeIdsStr: ", revokeIdsStr,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del crlIds
	_, err = session.Exec("delete from  lab_rpki_crl  where id in " + crlIdsStr)
	if err != nil {
		belogs.Error("delCrlDb():delete  from lab_rpki_crl fail: crlIdsStr: ", crlIdsStr,
			"   filePathPrefix:", filePathPrefix, "   err:", err)
		return err
	}
	belogs.Info("delCrlDb():delete lab_rpki_crl_*** ok, by filePathPrefix :", filePathPrefix,
		"  len(crlIds)", len(crlIds), "     time(s):", time.Now().Sub(start).Seconds())
	return nil

}
