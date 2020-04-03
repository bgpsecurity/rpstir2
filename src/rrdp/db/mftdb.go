package db

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/go-xorm/xorm"
)

func DelMft(session *xorm.Session, filePathPrefix string) (err error) {
	// try to delete old
	belogs.Debug("DelMft():will delete lab_rpki_mft by filePathPrefix :", filePathPrefix)

	mftIds := make([]int64, 0, 20000)
	err = session.SQL("select id from lab_rpki_mft Where filePath like ? ",
		filePathPrefix+"%").Find(&mftIds)
	if err != nil {
		belogs.Error("DelMft(): get mftIds fail, filePathPrefix: ", filePathPrefix, err)
		return err
	}
	belogs.Debug("DelMft():will delete lab_rpki_mft len(mftIds):", len(mftIds))

	for _, mftId := range mftIds {
		belogs.Debug("DelMft():delete lab_rpki_mft_ by mftId:", mftId)

		_, err := session.Exec("delete from lab_rpki_mft_file_hash  where mftId = ?", mftId)
		if err != nil {
			belogs.Error("DelMft():delete  from lab_rpki_mft_file_hash fail: mftId: ", mftId, err)
			return err
		}

		_, err = session.Exec("delete from  lab_rpki_mft_sia  where mftId = ?", mftId)
		if err != nil {
			belogs.Error("DelMft():delete  from lab_rpki_mft_sia fail:mftId: ", mftId, err)
			return err
		}

		_, err = session.Exec("delete from  lab_rpki_mft_aia  where mftId = ?", mftId)
		if err != nil {
			belogs.Error("DelMft():delete  from lab_rpki_mft_aia fail:mftId: ", mftId, err)
			return err
		}

		_, err = session.Exec("delete from  lab_rpki_mft  where id = ?", mftId)
		if err != nil {
			belogs.Error("DelMft():delete  from lab_rpki_mft fail:mftId: ", mftId, err)
			return err
		}
	}
	return nil

}
