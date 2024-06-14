package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"pagent/nic"
	"sync"
	"time"

	zmq "github.com/go-zeromq/zmq4"
	log "github.com/sirupsen/logrus"

	"pagent/utils"
)

var (
	arg_debug   = flag.Bool("D", false, "debug")
	arg_version = flag.Bool("v", false, "version")
	arg_list    = flag.Bool("l", false, "list devices")
	arg_config  = flag.String("f", "etc/pagentd.yaml", "config filename")
	START_TIME  = time.Now()
)

func init() {
	flag.Parse()
	if *arg_version {
		fmt.Printf("%s\n", utils.Version("pagentd"))
		os.Exit(0)
	}
	if *arg_list {
		nic.ListIfs()
		os.Exit(0)
	}

	utils.ShowBannerForApp("pagentd", utils.APP_VERSION, utils.BUILD_TIME)
	utils.Chdir2PrgPath()
	pwd, _ := utils.GetPrgDir()
	fmt.Println("pwd:", pwd)
	if err := utils.InitLog("pagentd.log", *arg_debug); err != nil {
		fmt.Printf("init log failed: %s\n", err)
		os.Exit(1)
	}
	log.Infof("BEGIN... %v, config=%v, debug=%v",
		START_TIME.Format(time.DateTime), *arg_config, *arg_debug)
}

func main() {
	// load config
	var myconfig, err = LoadConfig(*arg_config)
	if err != nil {
		log.Errorf("loadConfig error %s", err)
		os.Exit(1)
	}
	log.Infof("myconfig: %s", myconfig.Dump())

	var wg sync.WaitGroup
	chpkt := make(chan []byte, 1000000)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := zmq_pull(chpkt); err != nil {
			log.Printf("myserver: %s\n", err)
		}
	}()

	// to Devices name
	wg.Add(1)
	go func() {
		defer wg.Done()
		nic.WriteIf(myconfig.PcapOutput.Device, chpkt)
	}()

	wg.Wait()
}

func zmq_pull(chpkt chan []byte) error {
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
				fmt.Printf("receiving failed: %s\n", err)
				continue
			}
		}

		b := msg.Clone().Bytes()
		fmt.Printf("Received type:%d, count:%d, msglen:%d\n", msg.Type, count, len(b))

		if len(b) > 0 { // 最后一个包可能出现长度为0
			// fmt.Println(hex.Dump(b))

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
