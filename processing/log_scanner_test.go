package processing

import (
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestShortMultilineScan(t *testing.T) {
	const srcFileStr =
		`Line1
Line2
Line3
Line4
End`

	lineArr := []string{"End", "Line4", "Line3", "Line2", "Line1"}
	posArr := []int{24, 18, 12, 6, 0}
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