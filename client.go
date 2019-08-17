package main

import (
	"flag"
	//"fmt"
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
	size := 1400
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

// Header size is 3 bytes.
// 8-bit header has 2 fields:
//    FIN: 1 bit
//    Sequence number: 7 bits
//    File number: 16 bits // 0~65535
func addHeader(b [][]byte, fNum int) [][]byte {
	packets := make([][]byte, len(b))

	for i := 0; i < len(b); i++ {
		header := make([]byte, 3)
		header[0] = byte(i)
		header[1] = byte(fNum >> 8)  // Upper 8 bits
		header[2] = byte(fNum & 255) // Lower 8 bits
		if i == len(b)-1 {
			// Set a FIN flag.
			header[0] |= (1 << 7)
		}
		fNum2 := int(header[1]) << 8
		fNum2 |= int(header[2])
		//fmt.Println("seq, fnum:", int(header[0]), fNum, fNum2, header)
		packets[i] = append(header, b[i]...)
	}
	return packets
}

// Send all non-finished packets to a remote address.
func send(conn *net.UDPConn, packets [][][]byte, fNum int, fins [][]int32) {
	for {
		noSend := true
		for i := fNum; i < fNum+10; i++ {
			if i >= len(packets) {
				continue
			}
			for j := 0; j < 128; j++ {
				if j >= len(packets[i]) {
					return
				}
				if len(packets[i][j]) == 0 {
					return

				}
				//fmt.Println("SEND:", i, j, packets[i][j], len(packets), len(packets[i]))
				if atomic.LoadInt32(&fins[i][j]) == 0 {
					noSend = false
					conn.Write(packets[i][j])
				}
			}
		}
		if !noSend {
			return
		}
	}
}

func receive(conn *net.UDPConn, fins [][]int32) {
	buf := make([]byte, 1500)
	for {
		_, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
			//panic(err)
		}
		//fmt.Println("FIN:", buf[0])
		seq, _ := strconv.Atoi(string(buf[0]))
		fNum, _ := strconv.Atoi(string(buf[1:3]))
		//fmt.Println("Receive:", seq, fNum)
		atomic.StoreInt32(&fins[fNum][seq], 1)
		if checkSendAll(fins, fNum) {
			return
		}
		//fmt.Println("RECEIVE:", fins)
	}
}

func checkSendAll(fins [][]int32, fNum int) bool {
	for i := 0; i < len(fins[fNum]); i++ {
		if atomic.LoadInt32(&fins[fNum][i]) == 0 {
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
	i := 0
	nFile := 1000
	fPrefix := "../checkFiles/src/"

	// fins[fNum][Seq]
	// The elements of fins is an atomic variable.
	fins := make([][]int32, nFile)
	for j := 0; j < nFile; j++ {
		fins[j] = make([]int32, 128)
	}

	for {
		if i >= nFile - 11 {
			i = 0
		}
		packets := make([][][]byte, nFile)
		for j := 0; j < 128; j++ {
			packets[j] = make([][]byte, 128)
		}

		// Connect 10 files.
		for j := 0; j < 10; j++ {
			fNum := i + j
			raw := readfile(fPrefix + strconv.Itoa(fNum+1) + ".bin")
			bytes := split(raw)
			packets[j] = addHeader(bytes, fNum)
		}

		go receive(conn, fins)
		send(conn, packets, i, fins)
		i += 10 // Send 10 files at the same time.
	}
}
