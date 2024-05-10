package netif

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// capture packet from device to chan
func CapIf(device string, ch chan gopacket.Packet) error {
	if handle, err := pcap.OpenLive(device, 1600, true, pcap.BlockForever); err != nil {
		return err
	} else if err := handle.SetBPFFilter("tcp and port not 22"); err != nil { // optional
		return err
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for pkt := range packetSource.Packets() {
			// handle_pkt(pkt) // Do something with a packet here.
			ch <- pkt
		}
	}

	return nil
}
