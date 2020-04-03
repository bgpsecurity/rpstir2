package rtrtcp

import (
	"net"

	"rtr/rtrserver"
)

func RtrServerProcess(conn net.Conn) {
	rtrserver.RtrServerProcess(conn)
}
