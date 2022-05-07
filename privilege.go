package ping

import (
	"net"
	"sync"
)

var (
	PrivOnce   sync.Once
	NonPrivMsg string
	Privileged bool
)

func HasPrivilege() bool {
	PrivOnce.Do(func() {
		_, err := net.DialIP("ip4:icmp",
			&net.IPAddr{IP: net.ParseIP("0.0.0.0")},
			&net.IPAddr{IP: net.ParseIP("1.1.1.1")})
		if err != nil {
			Privileged = false
			NonPrivMsg = err.Error()
			return
		}
		Privileged = true
	})
	return Privileged
}

func init() {
	HasPrivilege()
}
