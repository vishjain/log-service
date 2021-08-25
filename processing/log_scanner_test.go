package processing

import (
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

// TestShortMultilineScan is a unit test which tests the Log Scanner over a
// basic multi-line input. You read small blocks of bytes at a time and check
// that the correct order of words were read and that the Seek is moving the file
// pointer to the right parts of the file.
func TestShortMultilineScan(t *testing.T) {
	const srcFileStr =
		`Line1
Line2 Second Word
Line3
Line4 Fourth Word
End`

	lineArr := []string{"End", "Line4 Fourth Word", "Line3", "Line2 Second Word", "Line1"}
	posArr := []int{48, 30, 24, 6, 0}
	var lineScanned []string
	var posScanned []int
	scanner := NewLogScanner(strings.NewReader(srcFileStr), len(srcFileStr), 5)
	for {
		line, pos, err := scanner.GetLine()
		if err != nil {
			assert.Equal(t, err, io.EOF)
			break
		}
		lineScanned = append(lineScanned, line)
		posScanned = append(posScanned, pos)
	}

	for idx, _ := range lineScanned {
		assert.Equal(t, lineScanned[idx], lineArr[idx])
	}

	for idx, _ := range posScanned {
		assert.Equal(t, posScanned[idx], posArr[idx])
	}
}
