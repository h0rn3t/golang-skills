package spool

import "io"

type lb struct {
	w   io.Writer
	buf []byte
}

// NewBuffer returns a writer that holds everything written to it until
// Render flushes it to w in one write.
func NewBuffer(w io.Writer) io.Writer {
	var b *lb
	b = &lb{}
	b.w = w
	b.buf = nil
	return b
}

func (b *lb) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *lb) fl() error {
	if len(b.buf) == 0 {
		return nil
	} else {
		_, err := b.w.Write(b.buf)
		b.buf = b.buf[:0]
		if err != nil {
			return err
		}
		return nil
	}
}
