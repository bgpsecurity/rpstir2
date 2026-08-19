package sync

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
)

// filePath, is nic dest path, eg: /home/rpki/data/reporrdp/rpki.apnic.cn/
func DelByFilePathDb(filePath string) (err error) {
	start := time.Now()
	belogs.Debug("DelByFilePathDb(): filePath:", filePath)
	if len(filePath) == 0 {
		belogs.Debug("DelByFilePathDb(): len(filePath) == 0:")
		return nil
	}

	session, err := xormdb.NewSession()
	if err != nil {
		belogs.Error("DelByFilePathDb(): NewSession fail :", err)
		return err
	}
	defer session.Close()

	err = delCerDb(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): delCerDb fail, filePath: ",
			filePath, err)
		return xormdb.RollbackAndLogError(session, "DelByFilePathDb(): delCerDb fail:", err)
	}

	err = delCrlDb(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): delCrlDb fail, filePath: ",
			filePath, err)
		return xormdb.RollbackAndLogError(session, "DelByFilePathDb(): delCrlDb fail:", err)
	}

	err = delMftDb(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): delMftDb fail, filePath: ",
			filePath, err)
		return xormdb.RollbackAndLogError(session, "DelByFilePathDb(): delMftDb fail:", err)
	}

	err = delRoaDb(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): delRoaDb fail, filePath: ",
			filePath, err)
		return xormdb.RollbackAndLogError(session, "DelByFilePathDb(): delRoaDb fail:", err)
	}

	err = delAsaDb(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): delAsaDb fail, filePath: ",
			filePath, err)
		return xormdb.RollbackAndLogError(session, "DelByFilePathDb(): delAsaDb fail:", err)
	}
	err = delMoaDb(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): delMoaDb fail, filePath: ",
			filePath, err)
		return xormdb.RollbackAndLogError(session, "DelByFilePathDb(): delMoaDb fail:", err)
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "DelByFilePathDb(): CommitSession fail:", err)
	}
	belogs.Info("DelByFilePathDb(): filePath:", filePath, "  time(s):", time.Since(start))

	return nil
}

// param: cerId/roaId/crlId/mftId
// paramIdsStr: cerIdsStr/roaIdsStr/crlIdsStr/mftIdsStr
func getIdsByParamIdsDb(tableName string, param string, paramIdsStr string) (ids []int64, err error) {
	belogs.Debug("getIdsByParamIdsDb():tableName:", tableName,
		"   param:", param, "   len(paramIdsStr):", len(paramIdsStr))
	ids = make([]int64, 0)
	// get ids from tableName
	err = xormdb.XormEngine.SQL("select id from " + tableName + " where " + param + " in " + paramIdsStr).Find(&ids)
	if err != nil {
		belogs.Error("getIdsByParamIdsDb(): get ids fail, tableName: ", tableName, "   param:", param,
			"  paramIdsStr:", paramIdsStr, err)
		return nil, err
	}

	belogs.Debug("getIdsByParamIdsDb():get ids from tableName: ", tableName, "   param:", param,
		"   len(paramIdsStr):", len(paramIdsStr), "  len(ids):", len(ids))
	return ids, nil
}
