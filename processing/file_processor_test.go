package processing

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// TestLastNEvents testes the file processor code path. It returns the last n events
// of a particular log file with a particular filter on.
func TestLastNEvents(t *testing.T) {
	fp := NewFileProcessor("test.log")

	// Open File & Call File Stat.
	file, fileInfo, err := fp.OpenFileAndStat("./")
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	// Close file descriptor.
	defer fp.FileClose()

	// Instantiate log scanner. We want to only count events that have the word
	// kernel.
	scanner := NewLogScanner(file, fileInfo.Size() - 1, 2048)
	filter := &FileFilter{
		noFilter: false,
		includeString: "kernel",
	}

	lineArr := []string{"Tue Aug 24 21:01:27.098 <kernel> installGTK: GTK installed",
		"Tue Aug 24 21:01:27.098 <kernel> GTK:"}

	var lineScanned []string

	maxLinesToRetrieve := 4

	// We want to retrieve a total of 5 events (with the condition that the line has kernel).
	// In this case, we don't have that many events (just 2) actually.
	totalNEvents := 5
	for {
		if totalNEvents <= 0 {
			break
		}

		linesToRetrieve := maxLinesToRetrieve
		if totalNEvents < linesToRetrieve {
			linesToRetrieve = totalNEvents
		}

		resp, err := fp.RetrieveNextFileEvents(scanner, filter, &RetrieveParams{
			maxLinesToRetrieve:linesToRetrieve,
		})
		assert.Nil(t, err)

		for _, line := range resp.lineList {
			lineScanned = append(lineScanned, line)
			totalNEvents -= 1
		}

		if resp.eof {
			break
		}
	}

	assert.Equal(t, len(lineScanned), len(lineArr))
	for idx, _ := range lineScanned {
		assert.Equal(t, lineScanned[idx], lineArr[idx])
	}
}