package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetMachineAddresses() []string {
	addresses := make([]string, 0)
	if hostname, err := os.Hostname(); err != nil {
		log.Errorf("fail to get hostname: %v", err)
	} else {
		addresses = append(addresses, hostname)
	}

	if ifAddresses, err := net.InterfaceAddrs(); err != nil {
		log.Errorf("fail to get ifs: %v", err)
	} else {
		for _, address := range ifAddresses {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				addresses = append(addresses, ipnet.IP.String())
			}
		}
	}

	return addresses
}

func BuildAccessURL(host string, port int, path string) url.URL {
	if strings.Count(host, ":") >= 2 { // ipv6
		host = fmt.Sprintf("[%v]", host)
	}
	host = fmt.Sprintf("%v:%v", host, port)
	url := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   path,
	}
	return url
}
