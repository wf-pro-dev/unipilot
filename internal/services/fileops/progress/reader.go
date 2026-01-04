package progress

import (
	"io"
)

// Reader wraps an io.Reader to track read progress
type Reader struct {
	reader  io.Reader
	tracker *Tracker
}

// NewReader creates a progress-tracking reader
func NewReader(reader io.Reader, tracker *Tracker) *Reader {
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
