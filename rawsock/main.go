package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket/layers"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	snapshot_len int32 = 65535
	promiscuous  bool  = false
	err          error
	timeout      time.Duration = 30 * time.Second
	handle       *pcap.Handle
	buffer       gopacket.SerializeBuffer
	options      gopacket.SerializeOptions
)

func main() {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	for _, value := range devices {
		if value.Description == "RZ608 Wi-Fi 6E 80MHz" {
			//Open device
			handle, err = pcap.OpenLive(value.Name, snapshot_len, promiscuous, timeout)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println(value.Description, value.Name)
	}
	// Send raw bytes over wire
	rawBytes := []byte{'A', 'b', 'C'}

	// This time lets fill out some information
	ipLayer := &layers.IPv4{
		Protocol: 17,
		Flags:    0x0000,
		IHL:      0x45,
		TTL:      0x80,
		Id:       0x1234,
		Length:   0x014e,
		SrcIP:    net.IP{0, 0, 0, 0},
		DstIP:    net.IP{255, 255, 255, 255},
	}
	ethernetLayer := &layers.Ethernet{
		EthernetType: 0x0800,
		SrcMAC:       net.HardwareAddr{0xFF, 0xAA, 0xFA, 0xAA, 0xFF, 0xAA},
		DstMAC:       net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}
	udpLayer := &layers.UDP{
		SrcPort: layers.UDPPort(68),
		DstPort: layers.UDPPort(67),
		Length:  0x013a,
	}
	// And create the packet with the layers
	buffer = gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(buffer, options,
		ethernetLayer,
		ipLayer,
		udpLayer,
		gopacket.Payload(rawBytes),
	)
	outgoingPacket := buffer.Bytes()
	for {
		time.Sleep(time.Second * 3)
		err = handle.WritePacketData(outgoingPacket)
		if err != nil {
			log.Fatal(err)
		}
	}

	handle.Close()
}
