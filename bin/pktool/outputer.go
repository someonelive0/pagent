package main

import (
	"os"
	"pagent/nic"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	log "github.com/sirupsen/logrus"
)

func output_file(chpkt chan []byte, stats *PktStats) error {
	var frame_len int

	// Open output pcap file and write header
	f, _ := os.Create("test.pcap")
	w := pcapgo.NewWriter(f)
	w.WriteFileHeader(65535, layers.LinkTypeEthernet)
	defer f.Close()
	pktOptions := gopacket.DecodeOptions{}
	pktOptions.Lazy = true
	pktOptions.NoCopy = true

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

		stats.Bytes += uint64(pcaphdr.Caplen)

		pkt := gopacket.NewPacket(frame[nic.PCAP_HDR_LEN:],
			layers.LayerTypeEthernet, pktOptions)
		pkt.Metadata().CaptureInfo.Timestamp = time.Unix(int64(pcaphdr.Sec), int64(pcaphdr.Usec)*1000)
		pkt.Metadata().CaptureInfo.Length = int(pcaphdr.Len)
		pkt.Metadata().CaptureInfo.CaptureLength = int(pcaphdr.Caplen)
		if pcaphdr.Len < pcaphdr.Caplen {
			pkt.Metadata().Truncated = true
		}
		// fmt.Println(pkt.Metadata())
		// nic.ParsePkt(pkt)

		if err := w.WritePacket(pkt.Metadata().CaptureInfo, pkt.Data()); err != nil {
			log.Errorf("write packate to file failed: %s", err)
		}

	}

	return nil
}
