package main

import (
	"context"
	"fmt"
	"io"

	zmq "github.com/go-zeromq/zmq4"
	log "github.com/sirupsen/logrus"
)

func zmq_init(port int) (zmq.Socket, error) {
	// Socket to talk to clients
	sock := zmq.NewPull(context.Background())
	addr := fmt.Sprintf("tcp://*:%d", port)
	if err := sock.Listen(addr); err != nil {
		return nil, fmt.Errorf("zmq listen addr %s failed: %w", addr, err)
	}
	log.Infof("zmq listen %s ok", addr)

	return sock, nil
}

func zmq_close(sock zmq.Socket) error {
	if sock != nil {
		return sock.Close()
	}
	return nil
}

func zmq_pull(sock zmq.Socket, chmsg chan []byte) error {

	count := 0
	for {
		msg, err := sock.Recv()
		if err != nil {
			if err == io.EOF { // EOF reached
				log.Infof("zmq recv EOF from %s", sock.Addr())
				continue
			} else if err == context.Canceled { // when zmqsock close
				break
			} else {
				log.Errorf("zmq recv failed: %s", err)
				return err
			}
		}

		b := msg.Clone().Bytes()
		log.Debugf("Received type:%d, count:%d, msglen:%d", msg.Type, count, len(b))

		if len(b) > 0 { // 按理包不会是0，但是以防万一
			// fmt.Println(hex.Dump(b))
			chmsg <- b
		} else {
			log.Warnf("zmq recv 0 len msg")
		}

		// fmt.Println("chmsg ", len(chmsg))
		count++

		// Do some 'work'
		// time.Sleep(time.Second)
	}

	return nil
}
