package sys

import (
	db "sys/db"
	sysmodel "sys/model"
)

func Results() (results sysmodel.Results, err error) {

	return db.Results()
}
