package main

import (
	"encoding/binary"

	log "github.com/sirupsen/logrus"

	"pagent/nic"
)

func worker(chmsg chan []byte, chpkt chan []byte, stats *PktStats) error {
	var msg_len, frame_len, offset int
	var i uint16

	for msg := range chmsg {
		msg_len = len(msg)
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

		offset = nic.BATCH_PKT_HDR_LEN // skip batch pkt hdr

		for i = 0; i < bathdr.PktsNum; i++ {
			if offset+2 > msg_len { // a short to show frame len
				break
			}
			frame_len = int(binary.BigEndian.Uint16(msg[offset:]))
			// log.Debugf("\tframe len:%d,\t", frame_len)
			offset += 2

			if offset+nic.PCAP_HDR_LEN+frame_len > msg_len {
				log.Warnf("frame len > msglen, %d, %d > %d", offset, frame_len, msg_len)
				break
			} else {
				chpkt <- msg[offset : offset+nic.PCAP_HDR_LEN+frame_len]
				offset += nic.PCAP_HDR_LEN + frame_len
			}

			// stats.Bytes += uint64(pcaphdr.Caplen), pcaphdr.Caplen should == frame_len
		}

	}

	return nil
}
