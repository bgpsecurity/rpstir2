package sync

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
)

// filePath, is nic dest path, eg: /root/rpki/data/reporrdp/rpki.apnic.cn/
func DelByFilePathDb(filePath string) (err error) {
	start := time.Now()
	belogs.Debug("DelByFilePathDb(): filePath:", filePath)
	if len(filePath) == 0 {
		belogs.Debug("DelByFilePathDb(): len(filePath) == 0:")
		return nil
	}

	session, err := xormdb.NewSession()
	defer session.Close()

	err = delCerDb(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): delCerDb fail, filePath: ",
			filePath, err)
		return err
	}

	err = delCrlDb(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): delCrlDb fail, filePath: ",
			filePath, err)
		return err
	}

	err = delMftDb(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): delMftDb fail, filePath: ",
			filePath, err)
		return err
	}

	err = delRoaDb(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): delRoaDb fail, filePath: ",
			filePath, err)
		return err
	}

	err = delAsaDb(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): delAsaDb fail, filePath: ",
			filePath, err)
		return err
	}

	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "DelByFilePathsDb(): CommitSession fail:", err)
	}
	belogs.Info("DelByFilePathsDb(): filePath:", filePath, "  time(s):", time.Since(start))

	return nil
}

// param: cerId/roaId/crlId/mftId
// paramIdsStr: cerIdsStr/roaIdsStr/crlIdsStr/mftIdsStr
func getIdsByParamIdsDb(tableName string, param string, paramIdsStr string) (ids []int64, err error) {
	belogs.Debug("getIdsByParamIdsDb():tableName :", tableName,
		"   param:", param, "   paramIdsStr:", paramIdsStr)
	ids = make([]int64, 0)
	// get ids from tableName
	err = xormdb.XormEngine.SQL("select id from " + tableName + " where " + param + " in " + paramIdsStr).Find(&ids)
	if err != nil {
		belogs.Error("getIdsByParamIdsDb(): get id fail, tableName: ", tableName, "   param:", param,
			"  paramIdsStr:", paramIdsStr, err)
		return nil, err
	}

	belogs.Debug("getIdsByParamIdsDb():get id fail, tableName: ", tableName, "   param:", param,
		"   paramIdsStr:", paramIdsStr, "  ids:", ids)
	return ids, nil
}
