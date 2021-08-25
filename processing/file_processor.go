package processing

import (
	"io"
	"os"
	"strings"
)

// FileProcessor holds information to perform file operations
// and to retrieve subsequent lines/data from the file.
type FileProcessor struct {
	file *os.File
	fileName string
	// TODO: Some sort of cache with the last n events (most recent probably)
	// of the file
	fileCache *FileCache
}

// TODO: FileCache holds events/lines from the particular file & additional
// metadata.
type FileCache struct {
}

// FileFilter describes whether a filter is being applied to retrieve the
// lines.
type FileFilter struct {
	noFilter bool
	// includeString is the string that must be present in the event.
	includeString string
}

type RetrieveParams struct {
	// maxLinesToRetrieve has the # of lines we need to return
	maxLinesToRetrieve int
}

type ResponseParams struct {
	// lineList is the slice of the line strings from file
	lineList []string
	// eof is a bool indicating if we have reached end of file.
	eof bool
}

func NewFileProcessor(fileName string) *FileProcessor {
	return &FileProcessor{
		fileName: fileName,
		fileCache: &FileCache{},
	}
}

// OpenFileAndStat will open relevant file.
func (fp *FileProcessor) OpenFileAndStat(baseDir string) (*os.File, os.FileInfo, error) {
	file, err := os.Open(baseDir + fp.fileName)
	if err != nil {
		return nil, nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	fp.file = file
	return file, fi, nil
}

// RetrieveNextFileEvents uses the log scanner to get # of lines (maxLinesToRetrieve)
// asked for from caller.
func (fp *FileProcessor) RetrieveNextFileEvents(scanner *LogScanner, filter *FileFilter,
	retrieveParams *RetrieveParams) (*ResponseParams, error) {
	var lineList []string

	for {
		// Using the log scanner, pull the next line. If we've already hit end-of-file,
		// return the lines we have. Otherwise, return the error.
		line, _, err := scanner.GetLine()
		if err != nil {
			if err == io.EOF {
				return &ResponseParams{
					lineList: lineList,
					eof: true,
				}, nil
			}
			return &ResponseParams{
				lineList: lineList,
				eof: false,
			}, err
		}

		// Check if there is a filter. If there is a filter, check if the returned string
		// continues the word we are looking for.
		if filter.noFilter {
			lineList = append(lineList, line)
		} else {
			if i := strings.Index(line, filter.includeString); i >= 0 {
				lineList = append(lineList, line)
			}
		}

		if len(lineList) == retrieveParams.maxLinesToRetrieve {
			return &ResponseParams{
				lineList: lineList,
				eof: false,
			}, nil
		}
	}
}

// FileClose will close relevant file.
func (fp *FileProcessor) FileClose() error {
	if fp.file != nil {
		return fp.file.Close()
	}
	return nil
}