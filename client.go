package main

import (
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
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		// Handle error.
	}
	defer conn.Close()

	fmt.Println("Send a message to server from client.")
	conn.Write([]byte("Hello From Client."))
	conn.Write([]byte("Hello From Client."))
	conn.Write([]byte("Hello From Client."))
	conn.Write([]byte("Hello From Client."))
}
