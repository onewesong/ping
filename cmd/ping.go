package main

import (
	"fmt"
	"os"
	"os/signal"
	"ping"
	"syscall"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug    = kingpin.Flag("debug", "Enable debug mode.").Bool()
	timeout  = kingpin.Flag("timeout", "Timeout waiting for ping in second.").Default("5s").Short('t').Duration()
	count    = kingpin.Flag("count", "Number of packets to send. default will be never end.").Default("-1").Short('c').Int()
	interval = kingpin.Flag("interval", "Interval of Ping").Default("1s").Short('i').Duration()
	localIp  = kingpin.Flag("local-ip", "Set local ip").Default("0.0.0.0").Short('l').IP()
	remoteIp = kingpin.Arg("ip", "IP address to ping.").Required().IP()
)

func main() {
	kingpin.Version("0.1.0")
	kingpin.Parse()
	if ping.Privileged != true {
		fmt.Println(ping.NonPrivMsg)
		os.Exit(1)
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	pinger := ping.NewPinger(localIp.String(), remoteIp.String(), *timeout, *count)
	pinger.Verbose = true
	pinger.OnFinish = func(stat *ping.Statistics) {
		fmt.Println("--- ping statistics ---")
		fmt.Printf("%+v\n", *stat)
	}
	go func() {
		<-c
		pinger.Finish()
		os.Exit(0)
	}()
	pinger.Run()
}
