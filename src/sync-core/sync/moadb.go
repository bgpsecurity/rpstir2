package sync

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/stringutil"
	"xorm.io/xorm"
)

func delMoaDb(session *xorm.Session, filePathPrefix string) (err error) {
	start := time.Now()
	belogs.Debug("delMoaDb():will delete lab_rpki_moa by filePathPrefix:", filePathPrefix)

	// get moaIds
	moaIds := make([]int64, 0)
	err = session.SQL("select id from lab_rpki_moa Where filePath like ? ",
		filePathPrefix+"%").Find(&moaIds)
	if err != nil {
		belogs.Error("delMoaDb(): get moaIds fail, filePathPrefix:", filePathPrefix, err)
		return err
	}
	if len(moaIds) == 0 {
		belogs.Debug("delMoaDb(): len(moaIds)==0, filePathPrefix:", filePathPrefix)
		return nil
	}
	moaIdsStr := stringutil.Int64sToInString(moaIds)
	belogs.Debug("delMoaDb():will delete lab_rpki_moa len(moaIds):", len(moaIds), "   filePathPrefix:", filePathPrefix)

	// del moaIds
	belogs.Debug("delMoaDb(): delete lab_rpki_moa, moaIdsStr:", moaIdsStr)
	_, err = session.Exec("delete from  lab_rpki_moa  where id in " + moaIdsStr)
	if err != nil {
		belogs.Error("delMoaDb():delete  from lab_rpki_moa fail: moaIdsStr:", moaIdsStr,
			"   filePathPrefix:", filePathPrefix, "   err:", err)
		return err
	}

	belogs.Info("delMoaDb():delete lab_rpki_moa ok, by filePathPrefix:", filePathPrefix,
		"  len(moaIds)", len(moaIds), "     time(s):", time.Since(start))
	return nil
}
