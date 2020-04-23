package sys

import (
	db "sys/db"
	sysmodel "sys/model"
)

func ExportRoas() (exportRoas []sysmodel.ExportRoa, err error) {

	return db.ExportRoas()
}
