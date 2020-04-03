package db

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/go-xorm/xorm"
)

func DelCrl(session *xorm.Session, filePathPrefix string) (err error) {
	// try to delete old
	belogs.Debug("DelCrl():will delete lab_rpki_crl by filePathPrefix :", filePathPrefix)

	crlIds := make([]int64, 0, 20000)
	err = session.SQL("select id from lab_rpki_crl Where filePath like ? ",
		filePathPrefix+"%").Find(&crlIds)
	if err != nil {
		belogs.Error("DelCrl(): get crlIds fail, filePathPrefix: ", filePathPrefix, err)
		return err
	}
	belogs.Debug("DelCrl():will delete lab_rpki_crl len(crlIds):", len(crlIds))

	for _, crlId := range crlIds {
		belogs.Debug("DelCrl():delete lab_rpki_crl_ by crlId:", crlId)

		_, err := session.Exec("delete from lab_rpki_crl_revoked_cert  where crlId = ?", crlId)
		if err != nil {
			belogs.Error("DelCrl():delete  from lab_rpki_crl_revoked_cert fail: crlId: ", crlId, err)
			return err
		}

		_, err = session.Exec("delete from  lab_rpki_crl  where id = ?", crlId)
		if err != nil {
			belogs.Error("DelCrl():delete  from lab_rpki_crl fail: crlId: ", crlId, err)
			return err
		}
	}
	return nil

}
