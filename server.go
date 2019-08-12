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

func analyze(packet []byte) (int, bool, []byte) {
	fin := false
	payload := make([]byte, len(packet)-1)
	copy(payload, packet[1:])    // Must copy because buffer is the same address.
	if packet[0]&(1<<7) == 128 { // 1000-0000b is 128
		fin = true
	}
	n := int(packet[0] & 127) // 0111-1111b is 127
	return n, fin, payload
}

func isRemaining(contents [][]byte, n int) bool {
	for i := 0; i < n; i++ {
		if len(contents[i]) == 0 {
			return true
		}
	}
	return false
}

func result(contents [][]byte, finSeq int) {
	fmt.Println("Result:")
	for i := 0; i < finSeq+1; i++ {
		fmt.Println(contents[i])
		fmt.Println(string(contents[i]))
	}
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
	contents := make([][]byte, 100)
	finSeq := -1
	for {
		n, client, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		//fmt.Print(buf[:n])
		fmt.Println("\n\nRecieved:", buf[:n], string(buf[:n]))

		// Analyze a header.
		seq, fin, payload := analyze(buf[:n])
		fmt.Println("HEADER:", seq, fin)
		fmt.Println("PAYLOAD:", payload)
		if fin {
			finSeq = seq
		}
		contents[seq] = payload[:]
		fmt.Println("CONTENT:", contents)

		// Send the sequence number to client.
		ans := []byte{byte(seq)}
		conn.WriteToUDP(ans, client)

		if finSeq != -1 {
			if isRemaining(contents, finSeq) {
				continue
			}
			result(contents, finSeq)
			return
		}
	}
}
