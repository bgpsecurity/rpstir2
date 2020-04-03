package db

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/go-xorm/xorm"
)

func DelRoa(session *xorm.Session, filePathPrefix string) (err error) {
	// try to delete old
	belogs.Debug("DelRoa():will delete lab_rpki_roa by filePathPrefix :", filePathPrefix)

	roaIds := make([]int64, 0, 20000)
	err = session.SQL("select id from lab_rpki_roa Where filePath like ? ",
		filePathPrefix+"%").Find(&roaIds)
	if err != nil {
		belogs.Error("DelRoa(): get roaIds fail, filePathPrefix: ", filePathPrefix, err)
		return err
	}
	belogs.Debug("DelRoa():will delete lab_rpki_roa len(roaIds):", len(roaIds))

	for _, roaId := range roaIds {
		belogs.Debug("DelRoa():delete lab_rpki_roa_ by roaId:", roaId)

		_, err := session.Exec("delete from lab_rpki_roa_ipaddress  where roaId = ?", roaId)
		if err != nil {
			belogs.Error("DelRoa():delete  from lab_rpki_roa_ipaddress fail: roaId: ", roaId, err)
			return err
		}

		_, err = session.Exec("delete from lab_rpki_roa_ee_ipaddress  where roaId = ?", roaId)
		if err != nil {
			belogs.Error("DelRoa():delete  from lab_rpki_roa_ee_ipaddress fail: roaId: ", roaId, err)
			return err
		}

		_, err = session.Exec("delete from  lab_rpki_roa_sia  where roaId = ?", roaId)
		if err != nil {
			belogs.Error("DelRoa():delete  from lab_rpki_roa_sia fail: roaId: ", roaId, err)
			return err
		}

		_, err = session.Exec("delete from  lab_rpki_roa_aia  where roaId = ?", roaId)
		if err != nil {
			belogs.Error("DelRoa():delete  from lab_rpki_roa_aia fail: roaId: ", roaId, err)
			return err
		}

		_, err = session.Exec("delete from  lab_rpki_roa  where id = ?", roaId)
		if err != nil {
			belogs.Error("DelRoa():delete  from lab_rpki_roa fail: roaId: ", roaId, err)
			return err
		}
	}
	return nil
}
