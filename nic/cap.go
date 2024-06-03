package nic

import (
	"fmt"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// capture packet from device to chan
func CapIf(device string, ch chan gopacket.Packet) error {
	handle, err := pcap.OpenLive(device, 65536, false, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	if err = handle.SetBPFFilter("tcp and port not 22"); err != nil { // optional
		return err
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for pkt := range packetSource.Packets() {
		// handle_pkt(pkt) // Do something with a packet here.
		ch <- pkt

		if len(ch) > 98 {
			fmt.Println("chan pkt is ", len(ch))
		}
	}

	return nil
}

// use libpcap level functions
func CapIf0(device string, ch chan gopacket.Packet) error {
	inHandle, err := pcap.NewInactiveHandle(device)
	if err != nil {
		return err
	}
	defer inHandle.CleanUp()

	if err = inHandle.SetSnapLen(65536); err != nil {
		return err
	}
	if err = inHandle.SetPromisc(false); err != nil {
		return err
	}
	if err = inHandle.SetTimeout(pcap.BlockForever); err != nil {
		return err
	}
	if err = inHandle.SetImmediateMode(true); err != nil { // packets are delivered to the application as soon as they arrive
		return err
	}
	if err = inHandle.SetBufferSize(30 * 1024 * 1024); err != nil {
		return err
	}

	timestampsrc := inHandle.SupportedTimestamps()
	fmt.Printf("SupportedTimestamps %#v\n", timestampsrc)
	for i := range timestampsrc {
		fmt.Printf("  SupportedTimestamp %d, %s\n", timestampsrc[i], timestampsrc[i].String())

	}
	// SupportedTimestamp 0, host
	// SupportedTimestamp 5, host_hiprec_unsynced
	// SupportedTimestamp 1, host_lowprec
	// SupportedTimestamp 2, host_hiprec
	inHandle.SetTimestampSource(0) // set timestamp source to host or host_lowprec

	handle, err := inHandle.Activate()
	if err != nil {
		return err
	}
	defer handle.Close()

	if err = handle.SetBPFFilter("tcp and port not 22"); err != nil { // optional
		return err
	}

	// such as: []pcap.Datalink{pcap.Datalink{Name:"EN10MB", Description:"Ethernet"}, pcap.Datalink{Name:"DOCSIS", Description:"DOCSIS"}}
	datalinks, err := handle.ListDataLinks()
	if err == nil {
		fmt.Printf("pcap datalinks %#v\n", datalinks)
	}

	drop_count := 0
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for pkt := range packetSource.Packets() {
		// handle_pkt(pkt) // Do something with a packet here.

		select {
		case ch <- pkt:
		default: // if chan is full then drop pkt
			fmt.Printf("chan pkt is full, drop pkt count %d\n", drop_count)
			drop_count++
			stats, err := handle.Stats()
			if err != nil {
				fmt.Println("pcap stat failed ", err)
			} else {
				fmt.Printf("pcap stats %#v\n", *stats)
			}
		}

		// if len(ch) > 98 {
		// 	fmt.Println("chan pkt is ", len(ch))
		// 	stats, err := handle.Stats()
		// 	if err != nil {
		// 		fmt.Println("pcap stat failed ", err)
		// 	} else {
		// 		fmt.Printf("pcap stats %#v\n", *stats)
		// 	}
		// }
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
