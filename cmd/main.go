package main

import (
	"errors"
	"fmt"
	processing "github.com/vishjain/log-service/processing"
	"net/http"
	"strconv"
)

const (
	// maxLinesToRetrieve is the # of lines that are written & flushed in 1 batch.
	maxLinesToRetrieve = 320
)

func main() {
	fileManager := processing.NewFileManager(maxLinesToRetrieve)

	// Listen and send events (blocks of read lines) to client via
	// invoking the handleLogQuery function.
	http.HandleFunc("/query", handleLogQuery(fileManager))
	http.ListenAndServe(":8001", nil)
}

// parseAndValidateQueryValues checks that the query parameters in the GET request from the user
// are valid.
func parseAndValidateQueryValues(r *http.Request) (*processing.QueryParams, error) {
	values := r.URL.Query()
	var fileName string
	var lastNEvents int
	var includeFilterStr string
	var err error


	// Must have a filename provided to query.
	fileNameLst, ok := values["filename"]
	if !ok {
		return nil, errors.New("No filename specified.")
	}

	if len(fileNameLst) != 1 {
		return nil, errors.New("Need exactly one file.")
	}
	fileName = fileNameLst[0]

	// Optional to provide n events you want to query.
	lastNEventsStrLst, ok := values["events"]
	if ok {
		if len(lastNEventsStrLst) != 1 {
			return nil, errors.New("Please provide how many events to see.")
		}
		lastNEventsStr := lastNEventsStrLst[0]
		lastNEvents, err = strconv.Atoi(lastNEventsStr)
		if err != nil {
			return nil, err
		}

		// Number must be greater or equal to 0.
		if lastNEvents < 0 {
			return nil, errors.New("Provide a positive number of events to see.")
		}
	}

	// Can include one word to filter by.
	includeFilterStrLst, ok := values["includefilter"]
	if ok {
		if len(includeFilterStrLst) != 1 {
			return nil, errors.New("Include filter 1 string.")
		}

		includeFilterStr = includeFilterStrLst[0]
	}

	return &processing.QueryParams{
		FileName: fileName,
		LastNEvents: lastNEvents,
		IncludeFilterStr: includeFilterStr,
	}, nil
}

// handleLogQuery is responsible for taking the relevant http request, validating it,
// kicking off the file reading, and sending the lines in the file to the client.
func handleLogQuery(fileManager *processing.FileManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Quick check on the input of the http request.
		queryParams, err := parseAndValidateQueryValues(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Subscribe channel.
		c := make(chan *processing.FileBlockReadInfo)

		// Spawn a goroutine to take the query parameters provided by client
		// and send block-by-block lines over the stream to the client.
		go func() {
			fileManager.ProcessLogQuery(c, queryParams)
		}()

		// Signal SSE Support
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")


		for {
			select {
			case <-r.Context().Done():
				return
			default:
				// Receive a block of read lines over a channel. Do some error check
				// before sending to client with http response writer.
				fileBlockReadInfo := <-c

				// Error out if there has been an error.
				if fileBlockReadInfo.Err != nil {
					http.Error(w, fileBlockReadInfo.Err.Error(), http.StatusInternalServerError)
					return
				}

				// Write all lines read & flush.
				for _, line := range fileBlockReadInfo.FileBlockRead {
					_, err := fmt.Fprintf(w, "%s\n", line)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				}
				w.(http.Flusher).Flush()
				if fileBlockReadInfo.FileProcessingFinished {
					return
				}
			}
		}
	}
}
