package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	udpAdr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Println("error creating the  connection", err)
		return
	}
	udpConn, err := net.DialUDP("udp", nil, udpAdr)
	if err != nil {
		fmt.Println("error creating the  connection", err)
		return
	}
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}
		byteSlice := []byte(line)
		_, err = udpConn.Write(byteSlice)
		if err != nil {
			fmt.Println("error writing to the byte", err)
			return
		}
	}
}
