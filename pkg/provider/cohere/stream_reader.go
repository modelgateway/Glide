package cohere

import (
	"bufio"
	"bytes"
	"context"
	"io"
)

// StreamReader reads Cohere streaming chat chunks that are formated
// as serializer chunk json per line (a.k.a. application/stream+json)
type StreamReader struct {
	scanner *bufio.Scanner
}

func containNewline(data []byte) (int, int) {
	return bytes.Index(data, []byte("\n")), 1
}

// NewStreamReader creates an instance of StreamReader
func NewStreamReader(stream io.Reader, maxBufferSize int) *StreamReader {
	scanner := bufio.NewScanner(stream)

	initBufferSize := min(4096, maxBufferSize)

	scanner.Buffer(make([]byte, initBufferSize), maxBufferSize)

	split := func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// We have a full event payload to parse.
		if i, nlen := containNewline(data); i >= 0 {
			return i + nlen, data[0:i], nil
		}

		// If we're at EOF, we have all the data.
		if atEOF {
			return len(data), data, nil
		}

		// Request more data.

		return 0, nil, nil
	}

	// Set the split function for the scanning operation.
	scanner.Split(split)

	return &StreamReader{
		scanner: scanner,
	}
}

// ReadEvent scans the EventStream for events.
func (r *StreamReader) ReadEvent() ([]byte, error) {
	if r.scanner.Scan() {
		event := r.scanner.Bytes()

		return event, nil
	}

	if err := r.scanner.Err(); err != nil {
		if err == context.Canceled {
			return nil, io.EOF
		}

		return nil, err
	}

	return nil, io.EOF
}
