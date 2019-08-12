package main

import (
	"flag"
	"fmt"
	"net"
)

func parseArgs() string {
	portPtr := flag.String("port", "8888", "The port number.")
	flag.Parse()
	return *portPtr
}

func main() {
	port := parseArgs()
	service := "0.0.0.0:" + port
	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Server is Running at " + conn.LocalAddr().String())
	buf := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		//fmt.Println("Received: ", string(buf[:n]), " from ", addr)
                fmt.Print(buf[:n])
	}
}
