package main

import (
	"context"
	"fmt"
	"sync"

	zmq "github.com/go-zeromq/zmq4"
	"github.com/google/gopacket"

	"pagent/netif"
)

func main() {
	netif.ListIf()
	fmt.Println("running...")

	var wg sync.WaitGroup
	chpkt := make(chan gopacket.Packet, 10000)
	ifname := "\\Device\\NPF_{27B6BF90-838D-43F0-AB4C-AAA823EF3285}"
	wg.Add(1)
	go func() {
		defer wg.Done()
		netif.CapIf(ifname, chpkt)
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

	count := 0
	for pkt := range chpkt {
		m := zmq.NewMsg(pkt.Data())
		m.Type = zmq.CmdMsg
		err := socket.Send(m)
		if err != nil {
			fmt.Printf("Send failed: %s\n", err)
			return err
		}
		count++
		// netif.HandlePkt(pkt)

		if count%10 == 0 {
			fmt.Println("---> ", count)
		}
	}

	return nil
}
