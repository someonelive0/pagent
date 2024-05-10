package main

import (
	"fmt"

	"github.com/google/gopacket/pcap"
)

func listif() ([]pcap.Interface, error) {
	// 得到所有的(网络)设备
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}
	// 打印设备信息
	fmt.Println("Devices found: ", len(devices))
	for _, device := range devices {
		fmt.Println("\nName: ", device.Name)
		fmt.Println("Description: ", device.Description)
		fmt.Println("Devices addresses: ", device.Description)
		fmt.Println("Flags: ", device.Flags)
		for _, address := range device.Addresses {
			fmt.Println("    - IP address: ", address.IP)
			fmt.Println("    - Subnet mask: ", address.Netmask)
		}
	}

	return devices, nil
}
