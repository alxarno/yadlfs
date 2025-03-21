package internal

import "io"

func newUploadFileProgress(r io.Reader, event Transfer, messages chan DialMessage) io.Reader {
	messages <- ProgressMessage{OID: event.OID, BytesSoFar: 0, BytesSinceLast: 0}

	progressCallback := func(bytesSoFar, bytesSinceLast int64) {
		messages <- ProgressMessage{OID: event.OID, BytesSoFar: bytesSoFar, BytesSinceLast: bytesSinceLast}
	}

	return newByteCountingReader(r, progressCallback)
}

func newDownloadFileProgress(w io.Writer, event Transfer, messages chan DialMessage) io.Writer {
	messages <- ProgressMessage{OID: event.OID, BytesSoFar: 0, BytesSinceLast: 0}

	progressCallback := func(bytesSoFar, bytesSinceLast int64) {
		messages <- ProgressMessage{OID: event.OID, BytesSoFar: bytesSoFar, BytesSinceLast: bytesSinceLast}
	}

	return newByteCountingWriter(w, progressCallback)
}
