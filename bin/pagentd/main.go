package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
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
			if err == io.EOF {
				// EOF reached
				fmt.Printf("receiving EOF %s\n", socket.Addr())
			} else {
				fmt.Printf("receiving: %s\n", err)
				continue
			}
		}

		b := msg.Clone().Bytes()
		fmt.Println("Received type ", msg.Type, count, len(b))

		if len(b) > 0 { // 最后一个包可能出现长度为0
			fmt.Println(hex.Dump(b))

			chpkt <- b
		}

		// fmt.Println("chpkt ", len(chpkt))
		count++

		// Do some 'work'
		// time.Sleep(time.Second)
	}
}

func toNic(device string, chpkt chan []byte) error {

	err := nic.WriteIf(device, chpkt)
	if err != nil {
		return err
	}

	return nil
}
