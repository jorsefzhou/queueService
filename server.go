package main

import (
	. "./queue"
	"fmt"
	"os"
	"flag"
)

// global variables
var listenPort   int
var bingdingIP   string
var passPeriod   int
var passNumber   int
var maxSeats     int
var helpMe       bool

// Set command line flags
func init() {
	// may change to listen eth number
	flag.StringVar(&bingdingIP, "ip", "127.0.0.1", "listen ip")
	flag.IntVar(&listenPort, "p", 6666, "listen port")
	flag.IntVar(&listenPort, "port", 6666, "listen port")
	flag.IntVar(&passPeriod, "t", 10, "seconds for one period")
	flag.IntVar(&passNumber, "c", 10, "token quata to dispatch in one period")
	flag.IntVar(&maxSeats, "max", 10000, "token quata to dispatch in one period")

	flag.BoolVar(&helpMe, "h", false, "help")
	flag.BoolVar(&helpMe, "help", false, "help")
}

func main() {
	flag.Parse()
	
	if helpMe {
		flag.PrintDefaults()
		os.Exit(0)
	}

	server := CreateServer(
		listenPort,
		bingdingIP,
		passPeriod,
		passNumber,
		maxSeats,
	)
	fmt.Printf("Running on %s\n", os.Args[1])
	server.Start()

}