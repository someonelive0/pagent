package main

import (
	"encoding/binary"

	log "github.com/sirupsen/logrus"

	"pagent/nic"
)

func worker(chmsg chan []byte, chpkt chan []byte, stats *PktStats) error {

	for msg := range chmsg {
		msg_len := len(msg)
		if msg_len < nic.BATCH_PKT_HDR_LEN {
			log.Warnf("zmq msg len %d < %d", msg_len, nic.BATCH_PKT_HDR_LEN)
			continue
		}
		// fmt.Println(hex.Dump(msg[:12+2+16]))
		stats.Msgs++

		bathdr, err := nic.BatchPktsHdrUnmarshal(msg)
		if err != nil {
			log.Errorf("BatchPktsHdrUnmarshal failed: %s", err)
		}
		log.Debugf("  batch hdr version:%d num:%d keybit:%d clientid:%d",
			bathdr.Version, bathdr.PktsNum, bathdr.KeyBit, bathdr.ClientId)

		if bathdr.PktsNum == 0 {
			continue
		}
		stats.Pkts += uint64(bathdr.PktsNum)

		offset := nic.BATCH_PKT_HDR_LEN // skip batch pkt hdr

		var i uint16 = 0
		for i = 0; i < bathdr.PktsNum; i++ {
			if offset+2 > msg_len { // a short to show frame len
				break
			}
			frame_len := binary.BigEndian.Uint16(msg[offset:])
			// log.Tracef("\tframe len:%d,\t", frame_len)
			offset += 2

			// get pcap-hdr
			if offset+nic.PCAP_HDR_LEN > msg_len {
				break
			}
			pcaphdr, err := nic.PcapHdrUnmarshal(msg[offset:])
			if err != nil {
				log.Errorf("PcapHdrUnmarshal failed: %s", err)
				break
			}
			offset += nic.PCAP_HDR_LEN

			log.Tracef("frame len:%d,\tpcaphdr sec:%d usec:%d caplen:%d len:%d",
				frame_len, pcaphdr.Sec, pcaphdr.Usec, pcaphdr.Caplen, pcaphdr.Len)
			offset += int(pcaphdr.Caplen)

			stats.Bytes += uint64(pcaphdr.Caplen)
		}

	}

	return nil
}
