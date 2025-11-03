package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func isValidHeaderKey(key string) bool {
	for _, char := range key {

		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}
	return len(key) > 0
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	crlfIndex := strings.Index(string(data), "\r\n")
	if crlfIndex == -1 {
		return 0, false, nil
	}

	if crlfIndex == 0 {
		return 2, true, nil
	}
	headerLine := string(data[:crlfIndex])

	colonIndex := strings.Index(headerLine, ":")
	if colonIndex == -1 {
		return 0, false, fmt.Errorf("invalid")
	}

	if strings.HasSuffix(headerLine[:colonIndex], " ") {
		return 0, false, fmt.Errorf("space before colonIndex")
	}

	key := strings.ToLower(strings.TrimSpace(headerLine[:colonIndex]))
	if !isValidHeaderKey(key) {
		return 0, false, fmt.Errorf("invalid character in header key")
	}
	value := strings.TrimSpace(headerLine[colonIndex+1:])
	if existing, exists := h[key]; exists {
		h[key] = existing + ", " + value
	} else {
		h[key] = value
	}

	return crlfIndex + 2, false, nil
}

func (h Headers) Get(key string) string {
	normalizedKey := strings.ToLower(key)
	return h[normalizedKey]
}
