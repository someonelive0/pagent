package nic

import (
	"encoding/binary"
	"fmt"

	"github.com/google/gopacket"
)

const PCAP_HDR_LEN = 16

// pcap header use big endian, header length is 16
type PcapHdr struct {
	Sec    uint32 /* timestamp seconds */
	Usec   uint32 /* timestamp microseconds */
	Caplen uint32 /* number of octets of packet saved in file */
	Len    uint32 /* actual length of packet */
}

func (p *PcapHdr) Marshal() []byte {
	var buf = make([]byte, PCAP_HDR_LEN)
	binary.BigEndian.PutUint32(buf, p.Sec)
	binary.BigEndian.PutUint32(buf[4:], p.Usec)
	binary.BigEndian.PutUint32(buf[8:], p.Caplen)
	binary.BigEndian.PutUint32(buf[12:], p.Len)
	return buf
}

func NewPcapHdr(capinfo *gopacket.CaptureInfo) *PcapHdr {
	return &PcapHdr{
		Sec:    uint32(capinfo.Timestamp.Unix()),
		Usec:   uint32(capinfo.Timestamp.Nanosecond() / 1000),
		Caplen: uint32(capinfo.CaptureLength),
		Len:    uint32(capinfo.Length),
	}
}

func PcapHdrUnmarshal(buf []byte) (*PcapHdr, error) {
	if len(buf) < PCAP_HDR_LEN {
		return nil, fmt.Errorf("unmarshal buffer length is less than %d", PCAP_HDR_LEN)
	}
	var hdr = &PcapHdr{
		Sec:    binary.BigEndian.Uint32(buf),
		Usec:   binary.BigEndian.Uint32(buf[4:]),
		Caplen: binary.BigEndian.Uint32(buf[8:]),
		Len:    binary.BigEndian.Uint32(buf[12:]),
	}

	// if hdr.Caplen > hdr.Len {
	// 	return nil, fmt.Errorf("unmarshal pcap header caplen bigger than len, %d > %d", hdr.Caplen, hdr.Len)
	// }
	return hdr, nil
}
