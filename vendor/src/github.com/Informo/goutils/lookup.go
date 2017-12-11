package goutils

import (
	"net"
	"strings"
)

func Lookup(serverName string) (success bool, fqdn string, port int) {
	cname, addrs, err := net.LookupSRV("matrix", "tcp", serverName)

	if cname == "" {
		println("Looked up SRV record for " + serverName)
	} else {
		println("Looked up SRV record for " + serverName + " (" + cname + ")")
	}

	if err != nil && !strings.HasSuffix(err.Error(), "no such host") {
		panic(err)
	} else if err != nil {
		return
	}

	success = true
	fqdn = addrs[0].Target
	port = int(addrs[0].Port)
	return
}
