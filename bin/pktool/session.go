package main

import (
	"encoding/json"
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
)

// tcp statistic
type TcpStats struct {
	Connections uint64 `json:"connections"` // tcp connections
	SyncConns   uint64 `json:"sync_conns"`  // tcp with sync connections
	HalfConns   uint64 `json:"half_conns"`  // tcp half capture connections

	TcpMap map[string]*TcpSession `json:"tcp_map"` // key is "sip dip"
	IpMap  map[string]*IpStats    `json:"ip_map"`
}

// a tcp session entry
type TcpSession struct {
	Sip     string
	Dip     string
	Bytes   uint64
	BytesRx uint64 `json:"bytes_rx"`
	BytesTx uint64 `json:"bytes_tx"`
}

// ip statistic
type IpStats struct {
	TcpConnections int    `json:"tcp_connections"`
	Bytes          uint64 `json:"bytes"`
	BytesRx        uint64 `json:"bytes_rx"`
	BytesTx        uint64 `json:"bytes_tx"`
}

func NewTcpStats() *TcpStats {
	tcpstats := &TcpStats{
		TcpMap: make(map[string]*TcpSession),
		IpMap:  make(map[string]*IpStats),
	}

	return tcpstats
}

func (p *TcpStats) Dump() []byte {
	b, _ := json.MarshalIndent(p, "", " ")
	return b
}

func (p *TcpStats) Count(pkt gopacket.Packet, ether *layers.Ethernet, ipLayer gopacket.Layer, tcp *layers.TCP) error {
	var sip, dip, tcpkey, tcpkey_r string // key is "sip:sport dip:dport"

	if ether.EthernetType == layers.EthernetTypeIPv4 { //EthernetTypeIPv4 = 0x0800
		ip, _ := ipLayer.(*layers.IPv4)
		sip = ip.SrcIP.String()
		dip = ip.DstIP.String()
	} else if ether.EthernetType == layers.EthernetTypeIPv6 { //EthernetTypeIPv6 = 0x86DD
		ip, _ := ipLayer.(*layers.IPv6)
		sip = ip.SrcIP.String()
		dip = ip.DstIP.String()
	}

	// process sip and dip
	if ip_stats, ok := p.IpMap[sip]; ok {
		ip_stats.Bytes += uint64(pkt.Metadata().CaptureLength)
		ip_stats.BytesTx += uint64(pkt.Metadata().CaptureLength)
	} else {
		ip_stats = &IpStats{
			Bytes:   uint64(pkt.Metadata().CaptureLength),
			BytesTx: uint64(pkt.Metadata().CaptureLength),
		}
		p.IpMap[sip] = ip_stats
	}

	if ip_stats, ok := p.IpMap[dip]; ok {
		ip_stats.Bytes += uint64(pkt.Metadata().CaptureLength)
		ip_stats.BytesRx += uint64(pkt.Metadata().CaptureLength)
	} else {
		ip_stats = &IpStats{
			Bytes:   uint64(pkt.Metadata().CaptureLength),
			BytesRx: uint64(pkt.Metadata().CaptureLength),
		}
		p.IpMap[dip] = ip_stats
	}

	// process tcp
	tcpkey = fmt.Sprintf("%s:%d %s:%d", sip, tcp.SrcPort, dip, tcp.DstPort)
	tcp_session, ok := p.TcpMap[tcpkey]
	if ok {
		tcp_session.Bytes += uint64(pkt.Metadata().CaptureLength)
		if tcp_session.Dip == dip {
			tcp_session.BytesRx += uint64(pkt.Metadata().CaptureLength)
		} else {
			tcp_session.BytesTx += uint64(pkt.Metadata().CaptureLength)
		}
	} else {
		tcp_session = &TcpSession{
			Sip:     sip,
			Dip:     dip,
			Bytes:   uint64(pkt.Metadata().CaptureLength),
			BytesTx: uint64(pkt.Metadata().CaptureLength),
		}
		p.TcpMap[tcpkey] = tcp_session
		tcpkey_r = fmt.Sprintf("%s:%d %s:%d", dip, tcp.DstPort, sip, tcp.SrcPort)
		p.TcpMap[tcpkey_r] = tcp_session

		log.Infof("new tcp session: %s, %s", tcpkey, tcpkey_r)
		if ip_stats, ok := p.IpMap[sip]; ok {
			ip_stats.TcpConnections++
		}
		if ip_stats, ok := p.IpMap[dip]; ok {
			ip_stats.TcpConnections++
		}

		p.Connections++
	}

	return nil
}
