package progress

import (
	"io"
	"net"
)

// Reader wraps an io.Reader to track read progress
type Reader struct {
	reader   io.ReadSeeker
	progress *Progress
}

type ConnReader struct {
	net.Conn
	OnWrite  func(n int64)
	progress *Progress
}

// NewReader creates a progress-tracking reader
func NewReader(reader io.ReadSeeker, progress *Progress) *Reader {
	return &Reader{
		reader:   reader,
		progress: progress,
	}
}

func (pr *Reader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)

	if n > 0 {
		pr.progress.Increment(int64(n))
	}

	return n, err
}

func (pr *Reader) Seek(offset int64, whence int) (int64, error) {
	return pr.reader.Seek(offset, whence)
}

func (cr *ConnReader) Write(p []byte) (int, error) {
	n, err := cr.Conn.Write(p)

	if n > 0 {
		cr.OnWrite(int64(n))
	}
	return n, err
}
