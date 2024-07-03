package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"pagent/nic"
	"pagent/utils"
)

var (
	arg_debug         = flag.Bool("D", false, "debug")
	arg_version       = flag.Bool("v", false, "version")
	arg_list          = flag.Bool("l", false, "list devices")
	arg_port          = flag.Int("p", 9266, "listen port")
	arg_if            = flag.String("i", "", "write packets to network interface name, such as eth0")
	arg_pcap_filename = flag.String("w", "", "write packets to pcap file name, such as pkts.pcap")
	arg_num           = flag.Int("n", 0, "recv pkts batch number, default 0 means no limit")
	START_TIME        = time.Now()
)

func init() {
	flag.Parse()
	if *arg_version {
		fmt.Printf("%s\n", utils.Version("pktool"))
		os.Exit(0)
	}
	if *arg_list {
		nic.ListIfs()
		os.Exit(0)
	}

	fmt.Printf("%s\n", utils.IDSS_BANNER)
	fmt.Printf("流量测试工具 pktool %s  Copyright (C) 2024 IDSS, Build on %s\n",
		utils.APP_VERSION, utils.BUILD_TIME)
	utils.Chdir2PrgPath()
	pwd, _ := utils.GetPrgDir()
	fmt.Println("pwd:", pwd)
	if err := utils.InitLog("pktool.log", *arg_debug); err != nil {
		fmt.Printf("init log failed: %s\n", err)
		os.Exit(1)
	}
	log.Infof("BEGIN... %v, port=%v, debug=%v",
		START_TIME.Format(time.DateTime), *arg_port, *arg_debug)
}

func main() {

	chmsg := make(chan []byte, 1000000)
	chpkt := make(chan []byte, 10000000)
	stats := PktStats{}
	var wg sync.WaitGroup

	zmqsock, err := zmq_init(*arg_port)
	if err != nil {
		log.Errorf("zmq init failed, %s", err)
		os.Exit(1)
	}

	err = output_init(arg_pcap_filename, arg_if)
	if err != nil {
		log.Errorf("output init failed, %s", err)
		os.Exit(1)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		output(chpkt, &stats)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		worker(chmsg, chpkt, &stats)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := zmq_pull(zmqsock, chmsg); err != nil {
			log.Errorf("zmq_pull failed: %s", err)
		}
	}()

	// timer
	timer_count := 0
	timer := func() {
		timer_count++
		if timer_count >= 10 {
			stats.CaculateSpeed()
			log.Infof("stats: %s\n", stats.Dump())
			timer_count = 0
		}
	}

	// GracefullExit, Wait and stop all
	gracefullExit := func() {
		log.Info("GracefullExit")
		zmq_close(zmqsock)
		close(chmsg)
		close(chpkt)
		wg.Wait()
		output_close()
	}

	loopUntilSignal(gracefullExit, timer)
}

// loop until signal such as SIGINT SIGTERM
func loopUntilSignal(gracefullExit func(), timer func()) {
	// Wait for signal and timer
	var signchan = make(chan os.Signal, 2)
	signal.Notify(signchan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	var ticker = time.NewTicker(time.Second * 1)

	toexit := func() {
		ticker.Stop()
		signal.Stop(signchan)
		close(signchan)

		gracefullExit()
	}

	// Waitting... signal and timer. block here.
	for {
		select {
		case <-ticker.C:
			if timer != nil {
				timer()
			}

		case s, ok := <-signchan:
			if !ok {
				goto END
			}
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				log.Info("Receive SIGNAL: ", s)
				toexit()
			default:
				log.Info("Receive other signal, ignore", s)
			}
		}
	}

END:
	log.Info("END")
}
