package main

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
