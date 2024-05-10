package main

import (
	"context"
	"fmt"
	"log"

	zmq "github.com/go-zeromq/zmq4"
)

func main() {
	for {
		if err := myserver(); err != nil {
			log.Printf("myserver: %s\n", err)
		}
	}
}

func myserver() error {
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
		fmt.Println("Received type ", msg.Type, count)

		// fmt.Println("Received ", msg.Bytes())
		count++

		// Do some 'work'
		// time.Sleep(time.Second)
	}
}
