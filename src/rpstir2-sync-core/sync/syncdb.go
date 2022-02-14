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

	err = DelCer(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): DelCer fail, filePath: ",
			filePath, err)
		return err
	}

	err = DelCrl(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): DelCrl fail, filePath: ",
			filePath, err)
		return err
	}

	err = DelMft(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): DelMft fail, filePath: ",
			filePath, err)
		return err
	}

	err = DelRoa(session, filePath)
	if err != nil {
		belogs.Error("DelByFilePathDb(): DelRoa fail, filePath: ",
			filePath, err)
		return err
	}
	err = xormdb.CommitSession(session)
	if err != nil {
		return xormdb.RollbackAndLogError(session, "DelByFilePathsDb(): CommitSession fail:", err)
	}
	belogs.Debug("DelByFilePathsDb(): filePath:", filePath, "  time(s):", time.Now().Sub(start).Seconds())

	return nil
}

// param: cerId/roaId/crlId/mftId
// paramIdsStr: cerIdsStr/roaIdsStr/crlIdsStr/mftIdsStr
func getIdsByParamIds(tableName string, param string, paramIdsStr string) (ids []int64, err error) {
	belogs.Debug("getIdsByParamIds():tableName :", tableName,
		"   param:", param, "   paramIdsStr:", paramIdsStr)
	ids = make([]int64, 0)
	// get ids from tableName
	err = xormdb.XormEngine.SQL("select id from " + tableName + " where " + param + " in " + paramIdsStr).Find(&ids)
	if err != nil {
		belogs.Error("getIdsByParamIds(): get id fail, tableName: ", tableName, "   param:", param,
			"  paramIdsStr:", paramIdsStr, err)
		return nil, err
	}

	belogs.Debug("getIdsByParamIds():get id fail, tableName: ", tableName, "   param:", param,
		"   paramIdsStr:", paramIdsStr, "  ids:", ids)
	return ids, nil
}
