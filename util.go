package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func getIPFromIPSB() []string {
	ips := make([]string, 0)
	client := &http.Client{}
	for _, url := range []string{"https://api-ipv4.ip.sb/ip", "https://api-ipv6.ip.sb/ip"} {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("User-Agent", "Mozilla")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("fail to read %v: %v", url, err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			ips = append(ips, strings.TrimSuffix(string(body), "\n"))
		}
	}
	return ips
}

func GetMachineAddresses(publicIPOnly bool) []string {
	addresses := make([]string, 0)
	if hostname, err := os.Hostname(); err != nil {
		log.Errorf("fail to get hostname: %v", err)
	} else {
		addresses = append(addresses, hostname)
	}

	if publicIPOnly {
		addresses = append(addresses, getIPFromIPSB()...)
		return addresses
	}

	if ifAddresses, err := net.InterfaceAddrs(); err != nil {
		log.Errorf("fail to get ifs: %v", err)
	} else {
		for _, address := range ifAddresses {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalMulticast() && !ipnet.IP.IsLinkLocalUnicast() {
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
