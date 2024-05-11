package main

import (
	"context"
	"fmt"
	"log"
	"pagent/nic"
	"sync"

	zmq "github.com/go-zeromq/zmq4"
)

func main() {

	var wg sync.WaitGroup
	chpkt := make(chan []byte, 10000)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := myserver(chpkt); err != nil {
			log.Printf("myserver: %s\n", err)
		}
	}()

	// to Devices addresses:  VirtualBox Host-Only Ethernet Adapter
	ifname := "\\Device\\NPF_{787AEC74-906E-45D7-AFE4-FCD4CF3E3F32}"
	wg.Add(1)
	go func() {
		defer wg.Done()
		nic.WriteIf(ifname, chpkt)
	}()

	wg.Wait()
}

func myserver(chpkt chan []byte) error {
	ctx := context.Background()
	// Socket to talk to clients
	socket := zmq.NewPull(ctx)
	defer socket.Close()
	if err := socket.Listen("tcp://*:9266"); err != nil {
		return fmt.Errorf("listening: %w", err)
	}
	fmt.Println("listen tcp://*:9266")

	count := 0
	for {
		msg, err := socket.Recv()
		if err != nil {
			fmt.Printf("receiving: %s\n", err)
		}

		b := msg.Clone().Bytes()
		fmt.Println("Received type ", msg.Type, count, len(b))

		if len(b) > 0 { // 最后一个包可能出现长度为0
			chpkt <- b
		}

		// fmt.Println("Received ", msg.Bytes())
		count++

		// Do some 'work'
		// time.Sleep(time.Second)
	}
}

// func tonic(device string, chpkt chan []byte) error {

// 	return nil
// }
