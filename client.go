package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

func parseArgs() (string, string, string) {
	var (
		hostPtr = flag.String("host", "localhost", "The host name.")
		portPtr = flag.String("port", "8888", "The port number.")
		filePtr = flag.String("file", "", "The file path for sending.")
	)
	flag.Parse()
	return *hostPtr, *portPtr, *filePtr
}

func readfile(f string) []byte {
	file, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	return b
}

// Split an original bytes to chunks which maximum size is 1400.
func split(raw []byte) [][]byte {
	var b [][]byte
	size := 14 // TODO: 1400
	for i := 0; ; i++ {
		if i*size > len(raw) {
			break
		}
		if i*size > len(raw)-size {
			// Last packet, so it should resize.
			remain := len(raw) - i*size
			b = append(b, raw[i*size:i*size+remain])
			break
		}
		b = append(b, raw[i*size:(i*size)+size])
	}
	return b
}

// 8-bit header has 2 fields:
//    FIN (1 bit)
//    Sequence number (7 bits)
func addHeader(b [][]byte) [][]byte {
	packets := make([][]byte, len(b))
	for i := 0; i < len(b); i++ {
		header := make([]byte, 1)
		header[0] = byte(i)
		if i == len(b)-1 {
			// Set a FIN flag.
			header[0] |= (1 << 7)
		}
		packets[i] = append(header, b[i]...)
	}
	return packets
}

// Send all non-finished packets to a remote address.
func send(conn *net.UDPConn, packets [][]byte, fins []bool) {
	for i := 0; i < len(packets); i++ {
		if !fins[i] {
			fmt.Println("SEND:", packets[i])
			conn.Write(packets[i])
		}
	}
}

func isRemaining(fins []bool) bool {
	for _, elm := range fins {
		if !elm {
			return true
		}
	}
	return false
}

func main() {
	host, port, file := parseArgs()
	service := host + ":" + port

	remoteAddr, err := net.ResolveUDPAddr("udp4", service)
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp4", nil, remoteAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	raw := readfile(file)
	fmt.Println("File content:", len(raw), raw)
	fmt.Println("File content:", string(raw))
	bytes := split(raw)
	packets := addHeader(bytes)
	fmt.Println(packets)

	fins := make([]bool, len(packets))

	buf := make([]byte, 1500)
	for isRemaining(fins) {
		send(conn, packets, fins)

		// This is an arbitrary number to wait reply.
		for i := 0; i < 10; i++ {
			_, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				panic(err)
			}
			fmt.Println("FIN:", buf[0])
			fins[int(buf[0])] = true
			if !isRemaining(fins) {
				return
			}
		}
	}
}
