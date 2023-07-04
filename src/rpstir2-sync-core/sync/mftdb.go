package sync

import (
	"time"

	"github.com/cpusoft/goutil/belogs"
	"github.com/cpusoft/goutil/xormdb"
	"xorm.io/xorm"
)

func delMftDb(session *xorm.Session, filePathPrefix string) (err error) {
	start := time.Now()
	belogs.Debug("delMftDb():will delete lab_rpki_mft_*** by filePathPrefix :", filePathPrefix)

	// get mftIds
	mftIds := make([]int64, 0)
	err = session.SQL("select id from lab_rpki_mft Where filePath like ? ",
		filePathPrefix+"%").Find(&mftIds)
	if err != nil {
		belogs.Error("delMftDb(): get mftIds fail, filePathPrefix: ", filePathPrefix, err)
		return err
	}
	if len(mftIds) == 0 {
		belogs.Debug("delMftDb(): len(mftIds)==0, filePathPrefix: ", filePathPrefix)
		return nil
	}
	mftIdsStr := xormdb.Int64sToInString(mftIds)
	belogs.Debug("delMftDb():will delete lab_rpki_mft len(mftIds):", len(mftIds), mftIdsStr,
		"   filePathPrefix:", filePathPrefix)

	// get filehashIds
	fileHashIds, err := getIdsByParamIdsDb("lab_rpki_mft_file_hash", "mftId", mftIdsStr)
	if err != nil {
		belogs.Error("delMftDb(): get fileHashIds fail, filePathPrefix: ", filePathPrefix,
			"   mftIdsStr:", mftIdsStr, err)
		return err
	}
	belogs.Debug("delMftDb(): len(fileHashIds):", len(fileHashIds), "   filePathPrefix:", filePathPrefix,
		"   mftIdsStr:", mftIdsStr)

	// get siaIds
	siaIds, err := getIdsByParamIdsDb("lab_rpki_mft_sia", "mftId", mftIdsStr)
	if err != nil {
		belogs.Error("delMftDb(): get siaIds fail, filePathPrefix: ", filePathPrefix,
			"   mftIdsStr:", mftIdsStr, err)
		return err
	}
	belogs.Debug("delMftDb(): len(siaIds):", len(siaIds), "   filePathPrefix:", filePathPrefix,
		"   mftIdsStr:", mftIdsStr)

	// get aiaIds
	aiaIds, err := getIdsByParamIdsDb("lab_rpki_mft_aia", "mftId", mftIdsStr)
	if err != nil {
		belogs.Error("delMftDb(): get aiaIds fail, filePathPrefix: ", filePathPrefix,
			"   mftIdsStr:", mftIdsStr, err)
		return err
	}
	belogs.Debug("delMftDb(): len(aiaIds):", len(aiaIds), "   filePathPrefix:", filePathPrefix,
		"   mftIdsStr:", mftIdsStr)

	// del filehashIds
	fileHashIdsStr := xormdb.Int64sToInString(fileHashIds)
	if len(fileHashIdsStr) > 0 {
		_, err := session.Exec("delete from lab_rpki_mft_file_hash  where id in " + fileHashIdsStr)
		if err != nil {
			belogs.Error("delMftDb():delete  from lab_rpki_mft_file_hash fail: fileHashIdsStr: ", fileHashIdsStr,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del siaIds
	siaIdsStr := xormdb.Int64sToInString(siaIds)
	if len(siaIdsStr) > 0 {
		_, err = session.Exec("delete from  lab_rpki_mft_sia  where id in " + siaIdsStr)
		if err != nil {
			belogs.Error("delMftDb():delete  from lab_rpki_mft_sia fail: siaIdsStr: ", siaIdsStr,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del siaIds
	aiaIdsStr := xormdb.Int64sToInString(aiaIds)
	if len(aiaIdsStr) > 0 {
		_, err = session.Exec("delete from  lab_rpki_mft_aia  where id in " + aiaIdsStr)
		if err != nil {
			belogs.Error("delMftDb():delete  from lab_rpki_mft_aia fail: aiaIdsStr: ", aiaIdsStr,
				"   filePathPrefix:", filePathPrefix, "   err:", err)
			return err
		}
	}

	// del mftIds
	_, err = session.Exec("delete from  lab_rpki_mft  where id in " + mftIdsStr)
	if err != nil {
		belogs.Error("delMftDb():delete  from lab_rpki_mft fail: mftIdsStr: ", mftIdsStr,
			"   filePathPrefix:", filePathPrefix, "   err:", err)
		return err
	}
	belogs.Info("delMftDb():delete lab_rpki_mft_*** ok, by filePathPrefix :", filePathPrefix,
		"  len(mftIds)", len(mftIds), "     time(s):", time.Since(start))
	return nil

}
