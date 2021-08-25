package processing

import (
	"io"
	"os"
	"strings"
)
type FileProcessor struct {
	file *os.File
	fileName string
	// Some sort of cache with the last n events (most recent probably)
	fileCache *FileCache
}

type FileCache struct {
}

type FileFilter struct {
	noFilter bool
	includeString string
}

type RetrieveParams struct {
	maxLinesToRetrieve int
}

type ResponseParams struct {
	lineList []string
	eof bool
}

func NewFileProcessor(fileName string) *FileProcessor {
	return &FileProcessor{
		fileName: fileName,
		fileCache: &FileCache{},
	}
}

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


func (fp *FileProcessor) RetrieveNextFileEvents(scanner *LogScanner, filter *FileFilter,
	retrieveParams *RetrieveParams) (*ResponseParams, error) {
	var lineList []string

	for {
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

func (fp *FileProcessor) FileClose() error {
	if fp.file != nil {
		return fp.file.Close()
	}
	return nil
}