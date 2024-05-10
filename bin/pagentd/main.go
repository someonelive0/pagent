package main

import (
	"context"
	"fmt"
	"log"

	zmq "github.com/go-zeromq/zmq4"
)

func main() {
	if err := hwserver(); err != nil {
		log.Fatalf("hwserver: %s\n", err)
	}
}

func hwserver() error {
	ctx := context.Background()
	// Socket to talk to clients
	socket := zmq.NewPull(ctx)
	defer socket.Close()
	if err := socket.Listen("tcp://*:9266"); err != nil {
		return fmt.Errorf("listening: %w", err)
	}

	for {
		msg, err := socket.Recv()
		if err != nil {
			return fmt.Errorf("receiving: %w", err)
		}
		fmt.Println("Received type ", msg.Type)

		fmt.Println("Received ", msg.Bytes())

		// Do some 'work'
		// time.Sleep(time.Second)
	}
}
