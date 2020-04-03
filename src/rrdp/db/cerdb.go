package db

import (
	belogs "github.com/astaxie/beego/logs"
	"github.com/go-xorm/xorm"
)

func DelCer(session *xorm.Session, filePathPrefix string) (err error) {
	belogs.Debug("DelCer():will delete lab_rpki_cer by filePathPrefix :", filePathPrefix)

	cerIds := make([]int64, 0, 20000)
	err = session.SQL("select id from lab_rpki_cer Where filePath like ? ",
		filePathPrefix+"%").Find(&cerIds)
	if err != nil {
		belogs.Error("DelCer(): get cerIds fail, filePathPrefix: ", filePathPrefix, err)
		return err
	}
	belogs.Debug("DelCer():will delete lab_rpki_cer len(cerIds):", len(cerIds))

	for _, cerId := range cerIds {
		belogs.Debug("DelCer():delete lab_rpki_cer_ by cerId:", cerId)

		_, err := session.Exec("delete from lab_rpki_cer_sia  where cerId = ?", cerId)
		if err != nil {
			belogs.Error("DelCer():delete  from lab_rpki_cer_sia failed, cerId:", cerId, err)
			return err
		}

		_, err = session.Exec("delete from  lab_rpki_cer_ipaddress  where cerId = ?", cerId)
		if err != nil {
			belogs.Error("DelCer():delete  from lab_rpki_cer_ipaddress failed, cerId:", cerId, err)
			return err
		}

		_, err = session.Exec("delete  from lab_rpki_cer_crldp  where cerId = ?", cerId)
		if err != nil {
			belogs.Error("DelCer():delete  from lab_rpki_cer_crldp failed, cerId:", cerId, err)
			return err
		}

		_, err = session.Exec("delete  from lab_rpki_cer_asn  where cerId = ?", cerId)
		if err != nil {
			belogs.Error("DelCer():delete  from lab_rpki_cer_asn  failed, cerId:", cerId, err)
			return err
		}

		_, err = session.Exec("delete  from lab_rpki_cer_aia  where cerId = ?", cerId)
		if err != nil {
			belogs.Error("DelCer():delete  from lab_rpki_cer_aia  failed, cerId:", cerId, err)
			return err
		}
		_, err = session.Exec("delete  from lab_rpki_cer  where id = ?", cerId)
		if err != nil {
			belogs.Error("DelCer():delete  from lab_rpki_cer  failed, cerId:", cerId, err)
			return err
		}

	}
	return nil
}
