package nic

import (
	"encoding/binary"
	"fmt"

	"github.com/google/gopacket"
)

// pcap header use little endian, header length is 16
type PcapHdr struct {
	Sec    uint32 /* timestamp seconds */
	Usec   uint32 /* timestamp microseconds */
	Caplen uint32 /* number of octets of packet saved in file */
	Len    uint32 /* actual length of packet */
}

func (p *PcapHdr) Marshal() []byte {
	var buf = make([]byte, 16)
	binary.LittleEndian.PutUint32(buf, p.Sec)
	binary.LittleEndian.PutUint32(buf[4:], p.Usec)
	binary.LittleEndian.PutUint32(buf[8:], p.Caplen)
	binary.LittleEndian.PutUint32(buf[12:], p.Len)
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
	if len(buf) < 16 {
		return nil, fmt.Errorf("unmarshal buffer length is less than 16")
	}
	var hdr = &PcapHdr{
		Sec:    binary.LittleEndian.Uint32(buf),
		Usec:   binary.LittleEndian.Uint32(buf),
		Caplen: binary.LittleEndian.Uint32(buf),
		Len:    binary.LittleEndian.Uint32(buf),
	}

	if hdr.Caplen > hdr.Len {
		return nil, fmt.Errorf("unmarshal pcap header caplen bigger than len, %d > %d", hdr.Caplen, hdr.Len)
	}
	return hdr, nil
}
