package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
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
	if n == -1 {
		return false
	}
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

func writeToFile(name string, contents [][]byte) {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for i := 0; i < len(contents); i++ {
		_, err := f.Write(contents[i])
		if err != nil {
			panic(err)
		}
	}
}

// RFC 768 for UDP
// https://tools.ietf.org/html/rfc768
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

	//fmt.Println("Server is Running at " + conn.LocalAddr().String())

	// Theoretical maximum size is 65,535 bytes (8 bytes header + 65,527 bytes payload).
	// Acutual maximum size of payload is 65,507 bytes
	// (= 65,535 bytes - 20 bytes IP header - 8 bytes header)

	// Settings for dst files.
	i := 1
	nFile := 2
	filePrefix := "./out/"
	//filePrefix := "../checkFiles/dst/"

	for {
		// Receive all packets for 1 file in this loop.
		finSeq := -1
		// 100 might not be enough.
		//contents := make([][]byte, 10000)
		contents := make([][]byte, 10)
	        buf := make([]byte, 1500)

		for finSeq == -1 || isRemaining(contents, finSeq) {
			// Receive a packet in this loop.
			// You need to execute this loop many times to complete a file.
			if i > nFile {
				fmt.Printf("Got %d files.\n", nFile)
				return
			}

			n, client, err := conn.ReadFromUDP(buf)
			if err != nil {
				panic(err)
			}
			fmt.Println("\n\nRecieved:", buf[:n], string(buf[:n]))

			// Analyze a header.
			seq, fin, payload := analyze(buf[:n])
			fmt.Println("HEADER:", seq, fin)
			fmt.Println("PAYLOAD:", payload)
			if fin {
				finSeq = seq
			}
			contents[seq] = payload
			fmt.Println("CONTENT:", contents)

			// Send the sequence number to client.
			ans := []byte{byte(seq)}
                        fmt.Println("ANS: ", ans)
			conn.WriteToUDP(ans, client)
		}

		writeToFile(filePrefix+strconv.Itoa(i)+".bin", contents)
		i++
	}
}
