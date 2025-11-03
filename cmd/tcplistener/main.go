package main

import (
	"fmt"
	"net"
	"os"

	"github.com/Kamausimon/httpFromTcp/internal/request"
)

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
			req, err := request.RequestFromReader(c)
			if err != nil {
				fmt.Println("there was an error trying to listen", err)
				return
			}
			fmt.Printf("- \nRequest line: \n - Method: %s\n, - Target: %s\n, - Version: %s\n",
				req.RequestLine.Method,
				req.RequestLine.RequestTarget,
				req.RequestLine.HttpVersion)
			os.Stdout.Sync()
			fmt.Println("Headers:")
			if len(req.Headers) == 0 {
				fmt.Println("- (no headers)")
			} else {
				for key, value := range req.Headers {
					fmt.Printf("- %s: %s\n", key, value)
				}
			}
			fmt.Println("Body:")
			fmt.Printf("\nBODY_STRING %s", req.Body)
		}(conn)
	}

}
