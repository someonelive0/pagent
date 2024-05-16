package nic

import (
	"fmt"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// capture packet from device to chan
func CapIf(device string, ch chan gopacket.Packet) error {
	if handle, err := pcap.OpenLive(device, 65536, false, pcap.BlockForever); err != nil {
		return err
	} else if err := handle.SetBPFFilter("tcp and port not 22"); err != nil { // optional
		return err
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for pkt := range packetSource.Packets() {
			// handle_pkt(pkt) // Do something with a packet here.
			ch <- pkt

			if len(ch) > 98 {
				fmt.Println("chan pkt is ", len(ch))
			}
		}
	}

	return nil
}

func WriteIf(device string, ch chan []byte) error {
	handle, err := pcap.OpenLive(device, 65536, true, pcap.BlockForever)
	if err != nil {
		log.Printf("OpenLive failed: %s", err)
		return err
	}

	for frame := range ch {
		err := handle.WritePacketData(frame)
		if err != nil {
			log.Printf("WritePacketData failed: %s", err)
		}
	}

	return nil
}
