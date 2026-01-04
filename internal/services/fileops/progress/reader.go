package progress

import (
	"io"
)

// Reader wraps an io.Reader to track read progress
type Reader struct {
	reader  io.ReadSeeker
	tracker *Tracker
}

// NewReader creates a progress-tracking reader
func NewReader(reader io.ReadSeeker, tracker *Tracker) *Reader {
	return &Reader{
		reader:  reader,
		tracker: tracker,
	}
}

func (pr *Reader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)

	if n > 0 {
		pr.tracker.Increment(int64(n))
	}

	return n, err
}

func (pr *Reader) Seek(offset int64, whence int) (int64, error) {
	return pr.reader.Seek(offset, whence)
}
