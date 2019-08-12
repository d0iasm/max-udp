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
