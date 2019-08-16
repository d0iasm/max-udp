package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync/atomic"
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
        //size := 1400
	size := 3
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
func send(conn *net.UDPConn, packets [][]byte, fins []int32) {
	for {
		isSend := false
		for i := 0; i < len(packets); i++ {
			if atomic.LoadInt32(&fins[i]) == 0 {
				isSend = true
				//fmt.Println("SEND:", packets[i])
				conn.Write(packets[i])
			}
		}
		if !isSend {
			return
		}
	}
}

func receive(conn *net.UDPConn, fins []int32) {
	finsNonAtom := make([]int32, len(fins))
	buf := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		//fmt.Println("FIN:", buf[0])
                i, _ := strconv.Atoi(string(buf[:n]))
                fmt.Println("Receive: ", i)
		atomic.StoreInt32(&fins[i], 1)
		finsNonAtom[i] = 1
        fmt.Println(finsNonAtom)
		if checkSendAll(finsNonAtom) {
			return
		}
	}
}

func checkSendAll(finsNonAtom []int32) bool {
	for i := 0; i < len(finsNonAtom); i++ {
		if finsNonAtom[0] == 0 {
			return false
		}
	}
	return true
}

// RFC 768 for UDP
// https://tools.ietf.org/html/rfc768
func main() {
	host, port, _ := parseArgs()
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

	// Settings for src files.
	i := 1
	nFile := 2
	filePrefix := "./"
	//filePrefix := "../checkFiles/src/"
	for {
		// Send all packets for 1 file in this loop.
		if i > nFile {
			return // TODO: Remove.
			i = 1
		}

		raw := readfile(filePrefix + strconv.Itoa(i) + ".bin")
		fmt.Println("File content:", strconv.Itoa(i)+".bin", len(raw), raw)
		fmt.Println("File content:", string(raw))
		bytes := split(raw)
		packets := addHeader(bytes)
		fmt.Println(packets)

		// The elements of fins is an atomic variable.
		fins := make([]int32, len(packets))

		go receive(conn, fins)
		send(conn, packets, fins)
		i++
	}
}
