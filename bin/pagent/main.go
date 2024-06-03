package main

import (
	"context"
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
			output(chpkt)
		}
	}()

	wg.Wait()
}

func output(chpkt chan gopacket.Packet) error {
	chszmq := make([]chan []byte, 3)
	chszmq[0] = make(chan []byte, 1)
	chszmq[1] = make(chan []byte, 1)
	chszmq[2] = make(chan []byte, 1)

	for i := range chszmq {
		go func(i int) {
			tozmq(chszmq[i], "tcp://127.0.0.1:9266")
		}(i)
	}

	count := 0
	for pkt := range chpkt {
		a, _ := pkt.Metadata().CaptureInfo.Timestamp.MarshalText()
		fmt.Printf("timestmp: %d, %d = %d  len %d, %d  - %s\n",
			pkt.Metadata().CaptureInfo.Timestamp.Unix(),
			pkt.Metadata().CaptureInfo.Timestamp.Nanosecond()/1000, // micro-second = nanosecond/1000
			pkt.Metadata().CaptureInfo.Timestamp.UnixMicro(),
			pkt.Metadata().CaptureInfo.CaptureLength, pkt.Metadata().CaptureInfo.Length, a)
		// var buf = make([]byte, 8)
		// binary.LittleEndian.PutUint32(buf, uint32(pkt.Metadata().CaptureInfo.Timestamp.Unix()))
		// binary.LittleEndian.PutUint32(buf[4:], uint32(pkt.Metadata().CaptureInfo.Timestamp.Nanosecond()/1000))
		// fmt.Println(hex.Dump(buf))

		pcaphdr := nic.NewPcapHdr(&pkt.Metadata().CaptureInfo)
		hdrbuf := pcaphdr.Marshal()
		frame := append(hdrbuf, pkt.Data()...) // default is ether frame. LINKTYPE_ETHERNET = 1

		count++
		// nic.HandlePkt(pkt)
		for i := range chszmq {
			select {
			case chszmq[i] <- frame:
			default:
				fmt.Printf("to chszmq failed %d\n", i)
			}
		}

		if count%10 == 0 {
			fmt.Println("---> ", count)
		}
	}

	return nil
}

func tozmq(chzmq chan []byte, addr string) error {
	ctx := context.Background()
	// Socket to talk to clients
	socket := zmq.NewPush(ctx)
	defer socket.Close()
	if err := socket.Dial(addr); err != nil {
		return fmt.Errorf("zmq dial %s failed: %w", addr, err)
	}
	fmt.Printf("connect %s ok", addr)
	socket.SetOption(zmq.OptionHWM, 1)

	for bs := range chzmq {
		m := zmq.NewMsg(bs)
		m.Type = zmq.CmdMsg
		err := socket.Send(m)
		if err != nil {
			fmt.Printf("Send failed: %s\n", err)
			return err
		}
	}

	return nil
}

/*
func tozmq0(chpkt chan gopacket.Packet) error {
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
*/
