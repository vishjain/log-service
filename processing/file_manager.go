package processing

import (
	"fmt"
)

type FileManager struct {
	fileToFileProcessor map[string]*FileProcessor
	maxLinesToRetrieve int
}


func NewFileManager(maxLinesToRetrieve int) *FileManager {
	return &FileManager{
		fileToFileProcessor: make(map[string]*FileProcessor),
		maxLinesToRetrieve: maxLinesToRetrieve,
	}
}

type QueryParams struct {
	FileName string
	LastNEvents int
	IncludeFilterStr string
}

type FileBlockReadInfo struct {
	FileBlockRead []string
	FileProcessingFinished bool
	Err error
}

func (fm *FileManager) ProcessLogQuery(c chan *FileBlockReadInfo, queryParams *QueryParams) {
	// Check if the file exists in map.
	var fileProcessor *FileProcessor
	var ok bool

	fileProcessor, ok = fm.fileToFileProcessor[queryParams.FileName]
	if !ok {
		fileProcessor = NewFileProcessor(queryParams.FileName)
	}

	// Open File & Call File Stat.
	file, fileInfo, err := fileProcessor.OpenFileAndStat("/var/log/")
	if err != nil {
		fmt.Printf("File Processor Error: %v \n", err)
		c <- &FileBlockReadInfo{
			Err: err,
		}
	}
	defer fileProcessor.FileClose()

	logScanner := NewLogScanner(file, int(fileInfo.Size()) - 1, 4096)
	if queryParams.LastNEvents > 0 {
		fm.processLastNEvents(fileProcessor, logScanner, c, queryParams)
	} else {
		fm.processWholeFile(fileProcessor, logScanner, c, queryParams)
	}

}


func (fm *FileManager) processWholeFile(fileProcessor *FileProcessor, logScanner *LogScanner,
	c chan *FileBlockReadInfo, queryParams *QueryParams) {

	filter := &FileFilter {
		noFilter: queryParams.IncludeFilterStr == "",
		includeString: queryParams.IncludeFilterStr,
	}

	for {
		resp, err := fileProcessor.RetrieveNextFileEvents(logScanner, filter, &RetrieveParams{
			maxLinesToRetrieve:fm.maxLinesToRetrieve,
		})
		if err != nil {
			fmt.Printf("processWholeFile Error: %v \n", err)
			c <- &FileBlockReadInfo{
				Err: err,
			}
			return
		}

		// Send block of lines read to server.
		c <- &FileBlockReadInfo{
			FileBlockRead:          resp.lineList,
			FileProcessingFinished: false,
			Err:                    nil,
		}

		// If the file has reached end of file, send to server.
		if resp.eof {
			fmt.Printf("End of file, processing complete. \n")
			c <- &FileBlockReadInfo{
				FileBlockRead:          nil,
				FileProcessingFinished: true,
				Err:                    nil,
			}
		}
	}
}

func (fm *FileManager) processLastNEvents(fileProcessor *FileProcessor, logScanner *LogScanner,
	c chan *FileBlockReadInfo, queryParams *QueryParams) {

	filter := &FileFilter {
		noFilter: queryParams.IncludeFilterStr == "",
		includeString: queryParams.IncludeFilterStr,
	}

	totalNEventsRemaining := queryParams.LastNEvents
	for {
		if totalNEventsRemaining <= 0 {
			c <- &FileBlockReadInfo{
				FileBlockRead:          nil,
				FileProcessingFinished: true,
				Err:                    nil,
			}
			return
		}

		linesToRetrieve := fm.maxLinesToRetrieve
		if totalNEventsRemaining < linesToRetrieve {
			linesToRetrieve = totalNEventsRemaining
		}

		resp, err := fileProcessor.RetrieveNextFileEvents(logScanner, filter, &RetrieveParams{
			maxLinesToRetrieve:linesToRetrieve,
		})
		if err != nil {
			fmt.Printf("processLastNEvents Error: %v \n", err)
			c <- &FileBlockReadInfo{
				Err: err,
			}
			return
		}

		// Send block of lines read to server.
		c <- &FileBlockReadInfo{
			FileBlockRead:          resp.lineList,
			FileProcessingFinished: false,
			Err:                    nil,
		}
		totalNEventsRemaining -= len(resp.lineList)

		// If the file has reached end of file, send to server.
		if resp.eof {
			fmt.Printf("End of file, processing complete. \n")
			c <- &FileBlockReadInfo{
				FileBlockRead:          nil,
				FileProcessingFinished: true,
				Err:                    nil,
			}
		}
	}

}