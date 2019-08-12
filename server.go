package main

import (
	"flag"
	"fmt"
	"net"
)

var port string
var host string

func parseArgs() {
	// String defines a string flag with specified name,
	// default value, and usage string.
	portPtr := flag.String("port", "8888", "The port number.")
	hostPtr := flag.String("host", "localhost", "The host name.")

	port = *portPtr
	host = *hostPtr
}

func main() {
	parseArgs()
	service := host + ":" + port

	udpAddr, _ := net.ResolveUDPAddr("udp", service)
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		// Handle error.
	}
	defer conn.Close()

	fmt.Println("Server is Running at " + service)
	buf := make([]byte, 1500)
	for {
		n, addr, _ := conn.ReadFromUDP(buf)
		fmt.Println("Received: ", string(buf[:n]), " from ", addr)
	}
}
