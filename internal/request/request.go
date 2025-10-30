package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
	State       State
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}
type State int

const (
	initialized State = iota
	done
)

func RequestFromReader(reader io.Reader) (*Request, error) {

	request := &Request{State: initialized}

	buffer := make([]byte, 0)

	readBuffer := make([]byte, 1024)

	for request.State != done {

		n, err := reader.Read(readBuffer)
		if err != nil && err != io.EOF {
			return nil, err
		}

		if n > 0 {

			buffer = append(buffer, readBuffer[:n]...)

			bytesConsumed, parseErr := request.parse(buffer)
			if parseErr != nil {
				return nil, parseErr
			}

			if bytesConsumed > 0 {
				buffer = buffer[bytesConsumed:]
			}
		}

		if err == io.EOF {
			break
		}
	}

	if request.State != done {
		return nil, fmt.Errorf("incomplete request")
	}

	return request, nil
}

func parseRequestLine(requestString string) (int, *Request, error) {
	crlfIndex := strings.Index(requestString, "\r\n")
	if crlfIndex == -1 {
		return 0, nil, nil
	}

	requestLine := requestString[:crlfIndex]
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return 0, nil, fmt.Errorf("the length is not accurate")
	}
	method := []rune(parts[0])
	for _, v := range method {
		if !unicode.IsUpper(v) || !unicode.IsLetter(v) {
			return 0, nil, fmt.Errorf("does not contain contain capital letters")
		}
	}
	http := parts[2]
	trimmed := strings.TrimPrefix(http, "HTTP/")
	if trimmed != "1.1" {
		return 0, nil, fmt.Errorf("http prefix does not contain 1.1")
	}
	bytesConsumed := crlfIndex + 2
	return bytesConsumed, &Request{
		RequestLine: RequestLine{
			Method:        string(method),
			HttpVersion:   trimmed,
			RequestTarget: parts[1],
		},
	}, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case initialized:
		bytesConsumed, request, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if bytesConsumed == 0 {
			return 0, nil
		}
		r.RequestLine = request.RequestLine
		r.State = done
		return bytesConsumed, nil

	case done:
		return 0, fmt.Errorf("error:trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error:unknown state")
	}
}
