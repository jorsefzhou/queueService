package main

import (
	. "./queue"
	"fmt"
	"os"
	"flag"
)


// Module globals
var concurrentClients int
var targetIp          string
var targetPort        int
var helpMe            bool
var reportStatistics  bool 


// Set command line flags
func init() {
	// Seed rand with current nanoseconds
	flag.IntVar(&concurrentClients, "c", 1, "concurrent clients")
	flag.StringVar(&targetIp, "ip", "127.0.0.1", "server ip")
	flag.IntVar(&targetPort, "p", 6666, "server port")
	flag.IntVar(&targetPort, "port", 6666, "server port")

	flag.BoolVar(&reportStatistics, "r", false, "report statistics")
	flag.BoolVar(&helpMe, "h", false, "help")
	flag.BoolVar(&helpMe, "help", false, "help")
}


func main() {
	flag.Parse()
	
	if helpMe {
		flag.PrintDefaults()
		os.Exit(0)
	}

	CreateCluster(
		concurrentClients,
		reportStatistics,
		targetIp,
		targetPort,
	)
	fmt.Printf("client started!\n0miao ")


}