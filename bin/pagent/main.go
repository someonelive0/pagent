package main

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"

	zmq "github.com/go-zeromq/zmq4"
	"github.com/google/gopacket"

	"pagent/nic"
)

func main() {
	nic.ListIfs()
	fmt.Println("running...")

	var wg sync.WaitGroup
	chpkt := make(chan gopacket.Packet, 100)

	// my local network interface
	ifname := "\\Device\\NPF_{27B6BF90-838D-43F0-AB4C-AAA823EF3285}"
	wg.Add(1)
	go func() {
		defer wg.Done()
		nic.CapIf0(ifname, chpkt)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			tozmq(chpkt)
		}
	}()

	wg.Wait()
}

func tozmq(chpkt chan gopacket.Packet) error {
	ctx := context.Background()
	// Socket to talk to clients
	socket := zmq.NewPush(ctx)
	defer socket.Close()
	if err := socket.Dial("tcp://127.0.0.1:9266"); err != nil {
		return fmt.Errorf("listening: %w", err)
	}
	fmt.Println("connect tcp://127.0.0.1:9266 ok")
	socket.SetOption(zmq.OptionHWM, 1)

	count := 0
	for pkt := range chpkt {
		a, _ := pkt.Metadata().CaptureInfo.Timestamp.MarshalText()
		fmt.Printf("timestmp: %d, %d = %d  len %d, %d  - %s\n",
			pkt.Metadata().CaptureInfo.Timestamp.Unix(),
			pkt.Metadata().CaptureInfo.Timestamp.Nanosecond()/1000, // micro-second = nanosecond/1000
			pkt.Metadata().CaptureInfo.Timestamp.UnixMicro(),
			pkt.Metadata().CaptureInfo.CaptureLength, pkt.Metadata().CaptureInfo.Length, a)
		var buf = make([]byte, 8)
		binary.LittleEndian.PutUint32(buf, uint32(pkt.Metadata().CaptureInfo.Timestamp.Unix()))
		binary.LittleEndian.PutUint32(buf[4:], uint32(pkt.Metadata().CaptureInfo.Timestamp.Nanosecond()/1000))
		fmt.Println(hex.Dump(buf))

		pcaphdr := nic.NewPcapHdr(&pkt.Metadata().CaptureInfo)
		hdrbuf := pcaphdr.Marshal()

		m := zmq.NewMsg(append(hdrbuf, pkt.Data()...))
		m.Type = zmq.CmdMsg
		err := socket.Send(m)
		if err != nil {
			fmt.Printf("Send failed: %s\n", err)
			return err
		}
		count++
		// nic.HandlePkt(pkt)

		if count%10 == 0 {
			fmt.Println("---> ", count)
		}
	}

	return nil
}
