package processing

import (
	"bytes"
	"io"
)

type LogScanner struct {
	r   io.ReaderAt
	pos int
	err error
	buf []byte
	blockSizeRead int
}

func NewLogScanner(r io.ReaderAt, pos int, blockSizeRead int) *LogScanner {
	return &LogScanner{r: r, pos: pos, blockSizeRead: blockSizeRead}
}

func (s *LogScanner) ReadMore() {
	// Check if we have reached the beginning of the file.
	if s.pos == 0 {
		s.err = io.EOF
		return
	}

	// Move s.pos back and read maximum of pre-specified blockSizeRead
	// bytes.
	sizeToRead := s.blockSizeRead
	if s.pos < sizeToRead {
		sizeToRead = s.pos
	}
	s.pos -= sizeToRead
	readBuffer := make([]byte, sizeToRead, sizeToRead + len(s.buf))

	_, s.err = s.r.ReadAt(readBuffer, int64(s.pos))
	if s.err == nil {
		s.buf = append(readBuffer, s.buf...)
	}
}

func (s *LogScanner) GetLine() (line string, start int, err error) {
	if s.err != nil {
		return "", 0, s.err
	}

	for {
		// Check if s.buf string has a new line character. Everything after new line character
		// is a line.
		lineBeginIdx := bytes.LastIndexByte(s.buf, '\n')
		if lineBeginIdx >= 0 {
			line, s.buf = string(s.buf[lineBeginIdx+1:]), s.buf[:lineBeginIdx]
			return line, s.pos + lineBeginIdx + 1, nil
		}

		// If there is no new line character, we need to read more.
		s.ReadMore()
		if s.err != nil {
			if s.err == io.EOF {
				if len(s.buf) > 0 {
					return string(s.buf), 0, nil
				}
			}
			return "", 0, s.err
		}
	}
}