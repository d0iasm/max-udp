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

func analyze(packet []byte) (int, bool, int, []byte) {
	fin := false
	payload := make([]byte, len(packet)-3)
	copy(payload, packet[3:]) // Must copy because buffer is the same address.

	// FIN.
	if packet[0]&(1<<7) == 128 { // 1000-0000b is 128
		fin = true
	}
	// Sequence number.
	n := int(packet[0] & 127) // 0111-1111b is 127
	// File number.
	fNum := int(packet[1]) << 8
	fNum |= int(packet[2])
	return n, fin, fNum, payload
}

func completeFile(fNum int, files [][][]byte, finSeqs []int) bool {
	if finSeqs[fNum] < 1 {
		return false
	}
	for i := 0; i < finSeqs[fNum]; i++ {
		if len(files[fNum][i]) == 0 {
			return false
		}
	}
	return true
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

	fmt.Println("Server is Running at " + conn.LocalAddr().String())

	// Theoretical maximum size is 65,535 bytes (8 bytes header + 65,527 bytes payload).
	// Acutual maximum size of payload is 65,507 bytes
	// (= 65,535 bytes - 20 bytes IP header - 8 bytes header)

	buf := make([]byte, 1500)

	// Settings for dst files.
	i := 0
	nFile := 1000
	fPrefix := "../checkFiles/dst/"

	// files[fNum][seq][i]
	//   fNum: File number.
	//   seq: Sequence number.
	files := make([][][]byte, nFile)
	finSeqs := make([]int, nFile)
	completed := make([]bool, nFile)

	// Initialize inner slice.
	for i := 0; i < nFile; i++ {
		files[i] = make([][]byte, 128)
	}

	for {
		if i > nFile {
			fmt.Printf("Got %d files.\n", nFile)
			return
		}

		n, client, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}
		fmt.Println("\n\nRecieved:", buf[:n], string(buf[:n]))

		if n < 1 {
			continue
		}

		// Analyze a header.
		seq, fin, fNum, payload := analyze(buf[:n])
		//fmt.Println("HEADER:", seq, fin, fNum)
		//fmt.Println("PAYLOAD:", payload)
		if fin {
			finSeqs[fNum] = seq
		}
		files[fNum][seq] = payload
		//fmt.Println("FILES:", files)

		// Reply has 2 fields:
		//   Sequence number: 7 bits (8 bits. The most upper bit is ignore.)
		//   File number: 16 bits
		ans := make([]byte, 3)
		ans[0] = byte(seq)
		ans[1] = byte(fNum >> 8)
		ans[2] = byte(fNum & 255)
		//fmt.Println("ANS: ", ans)
		_, err = conn.WriteToUDP(ans, client)
		if err != nil {
			panic(err)
		}

		if !completed[fNum] && completeFile(fNum, files, finSeqs) {
			fmt.Println("Get one file!", fNum)
			writeToFile(fPrefix+strconv.Itoa(fNum+1)+".bin", files[fNum])
			completed[fNum] = true
			i++
		}
	}

}
