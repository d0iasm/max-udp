package main

import (
	"fmt"
	"net"
)

func main() {
	host := "localhost"
	port := "33333"
	service := host + ":" + port

	udpAddr, _ := net.ResolveUDPAddr("udp", service)
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		// Handle error.
	}
	defer conn.Close()

	fmt.Println("Send a message to server from client.")
	conn.Write([]byte("Hello From Client."))
}
