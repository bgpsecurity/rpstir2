package parsevalidatedb

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"xorm.io/xorm"
)

func getCertIdByFilePathNameWithSession(session *xorm.Session, tableName, filePath, fileName string) (uint64, error) {
	belogs.Debug("getCertIdByFilePathNameWithSession(): tableName:", tableName, "  filePath:", filePath, "  fileName:", fileName)
	var id int
	sql := `select id from ` + tableName + ` where filepath=? and filename=?`
	has, err := session.SQL(sql, filePath, fileName).Get(&id)
	if err != nil {
		belogs.Error("getCertIdByFilePathNameWithSession(): get id failed, tableName:", tableName, " filePath:", filePath, "  fileName:", fileName, "    err:", err)
		return 0, err
	}
	if !has {
		belogs.Debug("getCertIdByFilePathNameWithSession(): not found from tableName:", tableName, " filePath:", filePath, "  fileName:", fileName)
		return 0, nil

	}
	belogs.Debug("getCertIdByFilePathNameWithSession(): get CertId, filePath:", filePath, "  fileName:", fileName, " id:", id)
	return uint64(id), nil
}

func selectByFilePathNameDbWithSession(session *xorm.Session, tableName, filePath, fileName, selectForUpdateWaitSec string) ([]int, error) {
	belogs.Debug("selectByFilePathNameDbWithSession(): tableName:", tableName, "  filePath:", filePath,
		"  fileName:", fileName, "  selectForUpdateWaitSec:", selectForUpdateWaitSec)

	start := time.Now()

	ids := make([]int, 0)
	sql := `select id from ` + tableName + ` where filepath=? and filename=? `
	belogs.Debug("selectByFilePathNameDbWithSession():select id from tableName, sql:", sql)
	err := session.SQL(sql, filePath, fileName).Find(&ids)
	if err != nil {
		belogs.Error("selectByFilePathNameDbWithSession(): select id failed, tableName:", tableName, " filePath:", filePath, "  fileName:", fileName,
			"    err:", err, " time(s):", time.Since(start))
		return nil, err
	}
	belogs.Debug("selectByFilePathNameDbWithSession(): pass select id, filePath:", filePath, "  fileName:", fileName,
		" ids:", ids, " time(s):", time.Since(start))
	return ids, nil
}
