package udt

import (
	"net"

	"github.com/murphybytes/ucp/udt/cudt"
)

// Dial connects to a UDT server
func Dial(addr string) (c net.Conn, e error) {
	var host, port string
	host, port, e = net.SplitHostPort(addr)

	if e != nil {
		return
	}

	var connectionKey int
	connectionKey, e = cudt.Dial(host, port)

	if e != nil {
		return
	}

	c = &Conn{
		connectionKey: connectionKey,
	}

	return
}
