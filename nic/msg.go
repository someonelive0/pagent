package nic

import (
	"encoding/binary"
	"fmt"
)

const BATCH_PKT_HDR_LEN = 12

// zmq message: BatchPktHeader follow with some PcapHdr and frames
type BatchPktsHdr struct {
	Version  uint16
	PktsNum  uint16
	KeyBit   uint32
	ClientId uint32
}

// it's network endian = big endian
func BatchPktsHdrUnmarshal(buf []byte) (*BatchPktsHdr, error) {
	if len(buf) < BATCH_PKT_HDR_LEN {
		return nil, fmt.Errorf("unmarshal buffer length is less than %d", BATCH_PKT_HDR_LEN)
	}
	var hdr = &BatchPktsHdr{
		Version:  binary.BigEndian.Uint16(buf),
		PktsNum:  binary.BigEndian.Uint16(buf[2:]),
		KeyBit:   binary.BigEndian.Uint32(buf[4:]),
		ClientId: binary.BigEndian.Uint32(buf[8:]),
	}

	return hdr, nil
}

/*
zmq message format, 32bit align

BatchPktsHdr(12 bytes), based on BatchPktsHdr.PktsNum
then loop get frame_len(2 bytes), then PCAP_HDR_LEN(16 bytes), then frame(frame_len)
means a frame = frame_len(2 bytes) + PCAP_HDR_LEN(16 bytes) + frame(frame_len)
and frame_len should equle PcapHdr.Caplen

+---------+---------+
| Version | PktsNum |
|      KeyBit       |
|     ClientId      |
| FrameLen|   Sec   |
| Sec     |   Usec  |
| Usec    | Caplen  |
| Caplen  | Len     |
| Len     | Frame...|
| FrameLen ...      |
+---------+---------+


*/
