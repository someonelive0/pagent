package main

import "github.com/google/gopacket"

func main() {
	listif()

	pktch := make(chan gopacket.Packet, 10000)
	ifname := "\\Device\\NPF_{27B6BF90-838D-43F0-AB4C-AAA823EF3285}"
	go capif(ifname, pktch)

	for pkt := range pktch {
		handle_pkt(pkt)
	}
}
