package processing

import (
	"fmt"
)

// FileManager is used to process query from the client.
type FileManager struct {
	// fileToFileProcessor maps the file path name to the
	// file processor.
	fileToFileProcessor map[string]*FileProcessor
	maxLinesToRetrieve int
}

func NewFileManager(maxLinesToRetrieve int) *FileManager {
	return &FileManager{
		fileToFileProcessor: make(map[string]*FileProcessor),
		maxLinesToRetrieve: maxLinesToRetrieve,
	}
}

// QueryParams holds information with the query parameters for the
// Get Request.
type QueryParams struct {
	FileName string
	LastNEvents int
	IncludeFilterStr string
}

// FileBlockReadInfo holds information about the lines
// just read from the file to be returned to the client.
type FileBlockReadInfo struct {
	// FileBlockRead holds the lines that were just read
	// from the file.
	FileBlockRead []string
	FileProcessingFinished bool
	// Err holds any error that occurred while processing
	// file.
	Err error
}

const (
	// blockSizeRead specifies how many bytes to retrieve in 1 file read call.
	blockSizeRead = 4096 * 16
	// directoryLocation is where we need to look for log files.
	directoryLocation = "/var/log/"
)

// ProcessLogQuery takes the query request and sends the relevant lines read to the http
// function handler (via a channel).
func (fm *FileManager) ProcessLogQuery(c chan *FileBlockReadInfo, queryParams *QueryParams) {
	// Check if the file exists in map.
	var fileProcessor *FileProcessor
	var ok bool

	fileProcessor, ok = fm.fileToFileProcessor[queryParams.FileName]
	if !ok {
		fileProcessor = NewFileProcessor(queryParams.FileName)
		fm.fileToFileProcessor[queryParams.FileName] = fileProcessor
	}

	// Open File & Call File Stat.
	file, fileInfo, err := fileProcessor.OpenFileAndStat(directoryLocation)
	if err != nil {
		fmt.Printf("File Processor Error: %v \n", err)
		c <- &FileBlockReadInfo{
			Err: err,
		}
		return
	}
	defer fileProcessor.FileClose()

	logScanner := NewLogScanner(file, fileInfo.Size() - 1, blockSizeRead)
	if queryParams.LastNEvents > 0 {
		fm.processLastNEvents(fileProcessor, logScanner, c, queryParams)
	} else {
		fm.processWholeFile(fileProcessor, logScanner, c, queryParams)
	}

}

// processWholeFile is a helper function which block by block sends read lines of a whole file
// over to the http server code. Applies relevant filters.
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

// processLastNEvents only sends the last n events over to the server handler code.
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