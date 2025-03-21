package internal

import (
	"fmt"
	"io"
)

type byteCountingWriter struct {
	writer   io.Writer
	callback func(bytesSoFar, bytesSinceLast int64)
	total    int64
}

func newByteCountingWriter(writer io.Writer, callback func(bytesSoFar, bytesSinceLast int64)) *byteCountingWriter {
	return &byteCountingWriter{
		writer:   writer,
		callback: callback,
		total:    0,
	}
}

func (bcw *byteCountingWriter) Write(p []byte) (int, error) {
	n, err := bcw.writer.Write(p)
	if n > 0 {
		bcw.total += int64(n)
		if bcw.callback != nil {
			bcw.callback(bcw.total, int64(n))
		}
	}

	return n, fmt.Errorf("failed to write: %w", err)
}

type byteCountingReader struct {
	reader   io.Reader
	callback func(bytesSoFar, bytesSinceLast int64)
	total    int64
}

func newByteCountingReader(reader io.Reader, callback func(bytesSoFar, bytesSinceLast int64)) *byteCountingReader {
	return &byteCountingReader{
		reader:   reader,
		callback: callback,
		total:    0,
	}
}

func (bcr *byteCountingReader) Read(p []byte) (int, error) {
	n, err := bcr.reader.Read(p)
	if n > 0 {
		bcr.total += int64(n)

		if bcr.callback != nil {
			bcr.callback(bcr.total, int64(n))
		}
	}

	return n, fmt.Errorf("failed to read: %w", err)
}
