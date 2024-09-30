package main

import (
	"fmt"
	"os"
	"pagent/nic"
	"sync/atomic"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	log "github.com/sirupsen/logrus"
)

var pcap_fp *os.File
var pcap_handle *pcap.Handle

func output_init(pcap_filename, net_device *string) error {

	if pcap_filename != nil && len(*pcap_filename) > 0 {
		f, err := os.Create(*pcap_filename)
		if err != nil {
			log.Errorf("create pcap file [%s] failed: %s", *pcap_filename, err)
			return err
		}
		pcap_fp = f
	}

	if net_device != nil && len(*net_device) > 0 {
		handle, err := pcap.OpenLive(*net_device, 65536, true, pcap.BlockForever)
		if err != nil {
			log.Printf("pcap OpenLive [%s] failed: %s", *net_device, err)
			pcap_fp.Close()
			return err
		}
		pcap_handle = handle
	}

	return nil
}

func output_close() error {
	if pcap_fp != nil {
		pcap_fp.Close()
		pcap_fp = nil
	}

	if pcap_handle != nil {
		pcap_handle.Close()
		pcap_handle = nil
	}

	return nil
}

func output(chpkt chan []byte, stats *PktStats, tcpstats *TcpStats) error {
	var frame_len int
	var w *pcapgo.Writer
	var pktOptions = gopacket.DecodeOptions{}
	pktOptions.Lazy = true
	pktOptions.NoCopy = true

	// Open output pcap file and write header
	// f, _ := os.Create("test.pcap")
	if pcap_fp != nil {
		w = pcapgo.NewWriter(pcap_fp)
		w.WriteFileHeader(65535, layers.LinkTypeEthernet)
	}
	// defer f.Close()

	for frame := range chpkt {
		frame_len = len(frame)

		// get pcap-hdr
		if nic.PCAP_HDR_LEN > frame_len {
			log.Warnf("frame len %d < PCAP_HDR_LEN %d", frame_len, nic.PCAP_HDR_LEN)
			continue
		}
		pcaphdr, err := nic.PcapHdrUnmarshal(frame)
		if err != nil {
			log.Errorf("PcapHdrUnmarshal failed: %s", err)
			continue
		}

		log.Debugf("frame len:%d,\tpcaphdr sec:%d usec:%d caplen:%d len:%d",
			frame_len, pcaphdr.Sec, pcaphdr.Usec, pcaphdr.Caplen, pcaphdr.Len)

		if frame_len != nic.PCAP_HDR_LEN+int(pcaphdr.Caplen) {
			log.Warnf("frame len is not equel nic.PCAP_HDR_LEN+caplen, %d != %d+%d",
				frame_len, nic.PCAP_HDR_LEN, pcaphdr.Caplen)
			continue
		}

		atomic.AddUint64(&stats.Bytes, uint64(pcaphdr.Caplen))

		pkt := gopacket.NewPacket(frame[nic.PCAP_HDR_LEN:],
			layers.LayerTypeEthernet, pktOptions)
		pkt.Metadata().CaptureInfo.Timestamp = time.Unix(int64(pcaphdr.Sec), int64(pcaphdr.Usec)*1000)
		pkt.Metadata().CaptureInfo.Length = int(pcaphdr.Len)
		pkt.Metadata().CaptureInfo.CaptureLength = int(pcaphdr.Caplen)
		if pcaphdr.Len < pcaphdr.Caplen {
			pkt.Metadata().Truncated = true
		}
		// fmt.Println(pkt.Metadata())

		process_pkt(pkt, tcpstats)

		if w != nil {
			if err := w.WritePacket(pkt.Metadata().CaptureInfo, pkt.Data()); err != nil {
				log.Errorf("write packate to file failed: %s", err)
			}
		}

		if pcap_handle != nil {
			err = pcap_handle.WritePacketData(pkt.Data())
			if err != nil {
				log.Errorf("pcap write packet (len %d) to net device failed: %s", pkt.Metadata().CaptureInfo.CaptureLength, err)
			}
		}

	}

	return nil
}

func process_pkt(pkt gopacket.Packet, tcpstats *TcpStats) error {
	// nic.ParsePkt(pkt)

	ethernetLayer := pkt.Layer(layers.LayerTypeEthernet)
	if ethernetLayer == nil {
		return fmt.Errorf("not ethernet packat")
	}
	ether, _ := ethernetLayer.(*layers.Ethernet)
	// Ethernet type is typically IPv4 but could be ARP or other
	fmt.Printf("  ether type %U, MAC: %s -> %s\n", ether.EthernetType, ether.SrcMAC, ether.DstMAC)

	var ipLayer gopacket.Layer
	var ipProtocol layers.IPProtocol                   //uint8
	if ether.EthernetType == layers.EthernetTypeIPv4 { //EthernetTypeIPv4 = 0x0800
		ipLayer = pkt.Layer(layers.LayerTypeIPv4)
		if ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)
			// IP layer variables:
			// Version (Either 4 or 6)
			// IHL (IP Header Length in 32-bit words)
			// TOS, Length, Id, Flags, FragOffset, TTL, Protocol (TCP?),
			// Checksum, SrcIP, DstIP
			fmt.Printf("  ip version %d, protocol %d, %s -> %s\n",
				ip.Version, ip.Protocol, ip.SrcIP, ip.DstIP)
			ipProtocol = ip.Protocol
		}
	} else if ether.EthernetType == layers.EthernetTypeIPv6 { //EthernetTypeIPv6 = 0x86DD
		ipLayer = pkt.Layer(layers.LayerTypeIPv6)
		if ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv6)
			// IP layer variables:
			// Version (Either 4 or 6)
			// IHL (IP Header Length in 32-bit words)
			// TOS, Length, Id, Flags, FragOffset, TTL, Protocol (TCP?),
			// Checksum, SrcIP, DstIP
			fmt.Printf("  ip version %d, protocol %d, %s -> %s\n",
				ip.Version, ip.NextHeader, ip.SrcIP, ip.DstIP)
			ipProtocol = ip.NextHeader
		}
	} else {
		return fmt.Errorf("unknown ethernet type %d", ether.EthernetType)
	}

	if ipProtocol == layers.IPProtocolTCP { // IPProtocol = 6
		tcpLayer := pkt.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			// TCP layer variables:
			// SrcPort, DstPort, Seq, Ack, DataOffset, Window, Checksum, Urgent
			// Bool flags: FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS
			fmt.Printf("  tcp seq %d, port %d -> %d\n", tcp.Seq, tcp.SrcPort, tcp.DstPort)
			tcpstats.Count(pkt, ether, ipLayer, tcp)
		} else {
			log.Warnf("not tcp")
		}

	} else if ipProtocol == layers.IPProtocolUDP { // IPProtocol = 17
		udpLayer := pkt.Layer(layers.LayerTypeUDP)
		if udpLayer != nil {
			udp, _ := udpLayer.(*layers.TCP)
			// TCP layer variables:
			// SrcPort, DstPort, Seq, Ack, DataOffset, Window, Checksum, Urgent
			// Bool flags: FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS
			fmt.Printf("  udp seq %d, port %d -> %d\n", udp.Seq, udp.SrcPort, udp.DstPort)
		} else {
			log.Warnf("not udp")
		}
	}

	return nil
}
