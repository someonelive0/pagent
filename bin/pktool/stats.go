package main

import (
	"encoding/json"
	"time"
)

// pkts statistic
type PktStats struct {
	Msgs     uint64 `json:"msgs"`      // zmq msgs total
	Pkts     uint64 `json:"pkts"`      // pcap frames total
	PktsReal uint64 `json:"pkts_real"` // pcap frames total
	Bytes    uint64 `json:"bytes"`     // pcap bytes total

	lastBytes uint64
	lastTimer time.Time
	Speed     uint64 `json:"speed"` // speed of pcap
}

func (p *PktStats) Dump() []byte {
	// b, _ := json.MarshalIndent(p, "", " ")
	b, _ := json.Marshal(p)
	return b
}

// Caculate speed Mbps MBps
func (p *PktStats) CaculateSpeed() uint64 {
	caculate_bytes := p.Bytes - p.lastBytes
	caculate_seconds := uint64(time.Since(p.lastTimer).Seconds())

	p.Speed = caculate_bytes / caculate_seconds

	p.lastBytes = p.Bytes
	p.lastTimer = time.Now()

	return 0
}
