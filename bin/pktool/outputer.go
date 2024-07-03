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

var pcap_fp *os.File

func output_init(pcap_filename, nic_device *string) error {

	if pcap_filename != nil && len(*pcap_filename) > 0 {
		f, err := os.Create(*pcap_filename)
		if err != nil {
			log.Errorf("create pcap file [%s] failed: %s", *pcap_filename, err)
			return err
		}
		pcap_fp = f
	}

	return nil
}

func output_close() error {
	if pcap_fp != nil {
		pcap_fp.Close()
		pcap_fp = nil
	}

	return nil
}

func output(chpkt chan []byte, stats *PktStats) error {
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

		if w != nil {
			if err := w.WritePacket(pkt.Metadata().CaptureInfo, pkt.Data()); err != nil {
				log.Errorf("write packate to file failed: %s", err)
			}
		}

	}

	return nil
}
