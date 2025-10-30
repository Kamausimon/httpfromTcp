package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		var allBytes []byte
		buffer := make([]byte, 8)
		for {
			n, err := f.Read(buffer)
			if n > 0 {
				allBytes = append(allBytes, buffer[:n]...)
			}
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Read error: %v\n", err)
				}
				break
			}

		}
		if len(allBytes) > 0 {
			content := string(allBytes)
			ch <- content
		}

		fmt.Print("\nconnection reading closed\n")
	}()
	return ch
}

func main() {

	In, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println("there was an error trying to listen", err)
		return
	}
	defer In.Close()

	for {
		conn, err := In.Accept()
		fmt.Print("connection made \n", conn)
		if err != nil {
			fmt.Println("there was an error connecting \n", err)
			continue
		}

		go func(c net.Conn) {
			defer c.Close()
			ch := getLinesChannel(c)
			for v := range ch {
				lines := strings.Split(v, "\n")
				for _, line := range lines {
					if line != "" {
						fmt.Printf("%s\n", line)
						os.Stdout.Sync()
					}
				}
			}
		}(conn)
	}

}
