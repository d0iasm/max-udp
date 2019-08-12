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

func analyze(packet []byte) (uint8, bool, []byte) {
	fin := false
	fmt.Println("HEADER:", packet[0], packet[0]&(1<<7), packet[0]&127)
	if packet[0]&(1<<7) == 128 { // 1000-0000b is 128
		fin = true
	}
	n := uint8(packet[0] & 127) // 0111-1111b is 127
	return n, fin, packet[1:]
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
	// 100 might not be enough.
	content := make([][]byte, 100)
	for {
		n, client, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		//fmt.Print(buf[:n])
		fmt.Println("Recieved:", buf[:n], string(buf[:n]))
		// Send the sequence number to client for set a FIN field.
		seq, fin, payload := analyze(buf[:n])
		ans := []byte{byte(seq)}
		conn.WriteToUDP(ans, client)
		content[seq] = payload
		if fin {
			fmt.Println("Get final packet.")
		}
	}
}
