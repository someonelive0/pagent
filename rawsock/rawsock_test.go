package main

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestInterfaceAddrs(t *testing.T) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n\n", addrs)
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Printf("%s   %s\n", ipnet.IP.String(), ipnet.Network())
			}
		}
	}
}

func TestInterfaces(t *testing.T) {
	ifs, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%#v\n\n", ifs)

	for _, inf := range ifs {
		fmt.Printf("name=%s   mac=%s\n", inf.Name, inf.HardwareAddr)
	}
}

func TestGetOutBoundIP(t *testing.T) {
	conn, err := net.Dial("udp", "192.168.30.59:53")
	if err != nil {
		fmt.Println(err)
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr.String())
	ip := strings.Split(localAddr.String(), ":")[0]
	fmt.Printf("out bound ip: %s\n", ip)
}
