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
	size := 1400
	for i := 0; ; i++ {
		if i*size > len(raw) {
			break
		}
		if i*size > len(raw)-size {
			// Last packet, so it should resize.
                        remain := len(raw) - i*size
			b = append(b, raw[i*size:i*size + remain])
			break
		}
		b = append(b, raw[i*size:(i*size)+size])
	}
	return b
}

func addHeader() {
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
