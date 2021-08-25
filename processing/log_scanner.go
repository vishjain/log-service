package processing

import (
	"bytes"
	"io"
)

// LogScanner is used to read from the end of the file. Handles
// internal details of reading the file.
type LogScanner struct {
	r   io.ReaderAt
	// pos represents the file pointer position to do file reading.
	pos int
	// err is the error after a file operations.
	err error
	// buf contains the bytes just read from file.
	buf []byte
	// blockSizeRead is the amount you read from file at one time.
	blockSizeRead int
}

func NewLogScanner(r io.ReaderAt, pos int, blockSizeRead int) *LogScanner {
	return &LogScanner{r: r, pos: pos, blockSizeRead: blockSizeRead}
}

// ReadMore is called to read more bytes from file into buffer if possible.
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

// GetLine returns the next line string in the file. It is meant to be
// successively called to get the next line.
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