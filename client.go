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

// Split an original bytes to N chunks.
func split(raw []byte) [][]byte {
	n := 10
	leng := len(raw) / (n - 1)
	b := make([][]byte, n)
	for i := 0; i < n-1; i++ {
		b[i] = raw[i*leng : (i*leng)+leng]
	}
	b[n-1] = raw[leng*(n-1):]
	return b
}

func main() {
	host, port, file := parseArgs()
	service := host + ":" + port

	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	raw := readfile(file)
	fmt.Println("File content: ", len(raw), raw)
	fmt.Println("File content: ", string(raw))
	bytes := split(raw)

	fmt.Println("Send a message to server from client.")
	for i := 0; i < len(bytes); i++ {
		header := make([]byte, 1)
		header[0] = byte(i)
		conn.Write(append(header, bytes[i]...))
	}
}
