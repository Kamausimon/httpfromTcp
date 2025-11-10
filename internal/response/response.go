package response

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Kamausimon/httpFromTcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
	writerStateTrailers
)

type Writer struct {
	writer      io.Writer
	writerState writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writerState: writerStateStatusLine,
		writer:      w}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("cannot write status line in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateHeaders }()
	var statusText string
	switch statusCode {
	case StatusOK:
		statusText = "OK"
	case StatusBadRequest:
		statusText = "Bad Request"
	case StatusInternalServerError:
		statusText = "Internal Server Error"
	default:
		statusText = "Unknown Status"
	}
	_, err := fmt.Fprintf(w.writer, "HTTP/1.1 %d %s \r\n", statusCode, statusText)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["content-length"] = fmt.Sprintf("%d", contentLen)

	h["connection"] = "close"

	h["content-type"] = "text/html"

	return h

}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("cannot write headers in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateBody }()
	for key, value := range headers {
		_, err := fmt.Fprintf(w.writer, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}

	_, err := io.WriteString(w.writer, "\r\n")
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	const defaultChunkSize = 1024

	totalWritten := 0

	for i := 0; i < len(p); i += defaultChunkSize {
		end := i + defaultChunkSize
		if end > len(p) {
			end = len(p)
		}
		chunk := p[i:end]
		chunklen := len(chunk)

		_, err := fmt.Fprintf(w.writer, "%x\r\n", chunklen)
		if err != nil {
			return totalWritten, err
		}
		n, err := w.writer.Write(chunk)
		totalWritten += n
		if err != nil {
			return totalWritten, err
		}

		_, err = io.WriteString(w.writer, "\r\n")
		if err != nil {
			return totalWritten, err
		}
		if flusher, ok := w.writer.(interface{ Flush() error }); ok {
			flusher.Flush()
		}
		time.Sleep(50 * time.Millisecond)

	}
	return totalWritten, nil
}
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	n, err := w.writer.Write([]byte("0\r\n"))
	if err != nil {
		return n, err
	}
	w.writerState = writerStateTrailers
	return n, nil
}

func (w *Writer) WriteChunkedBodyDoneWithoutTrailers() error {
	_, err := io.WriteString(w.writer, "0\r\n\r\n")
	if err != nil {
		return err
	}
	if flusher, ok := w.writer.(interface{ Flush() error }); ok {
		flusher.Flush()
	}
	return nil
}

func GetChunkedHeaders() headers.Headers {
	h := headers.NewHeaders()
	h["transfer-encoding"] = "chunked"
	h["connection"] = "close"
	h["content-type"] = "text/plain"
	return h
}

func GetChunkedHeadersWithTrailers(trailerNames []string) headers.Headers {
	h := headers.NewHeaders()
	h["transfer-encoding"] = "chunked"
	h["connection"] = "close"
	h["content-type"] = "text/plain"

	// Announce which headers will be in trailers
	if len(trailerNames) > 0 {
		h["trailer"] = strings.Join(trailerNames, ", ")
	}

	return h
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writerState != writerStateTrailers {
		return fmt.Errorf("cannot write trailers in state %d", w.writerState)
	}
	defer func() { w.writerState = writerStateBody }()
	for k, v := range h {
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}
