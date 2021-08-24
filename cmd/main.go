package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func main() {
	nc := NewNotificationCenter()

	go func() {
		for {
			b := []byte(time.Now().Format(time.RFC3339))
			if err := nc.Notify(b); err != nil {
				log.Fatal(err)
			}

			time.Sleep(1 * time.Second)
		}
	}()

	http.HandleFunc("/query", handleLogQuery(nc))
	http.ListenAndServe(":8001", nil)
}

type UnsubscribeFunc func() error

type Subscriber interface {
	Subscribe(c chan []byte) (UnsubscribeFunc, error)
}
type QueryParams struct {
	fileName string
	lastNEvents int
	includeFilterStr string
}

func parseAndValidateQueryValues(r *http.Request) (*QueryParams, error) {
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

		// Number must be greater than 0.
		if lastNEvents <= 0 {
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

	return &QueryParams{
		fileName: fileName,
		lastNEvents: lastNEvents,
		includeFilterStr: includeFilterStr,
	}, nil
}

func handleLogQuery(s Subscriber) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Quick check on the input of the http request.
		queryParams, err := parseAndValidateQueryValues(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Printf("Query params %+v \n", *queryParams)

		// Subscribe
		c := make(chan []byte)
		unsubscribeFn, err := s.Subscribe(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Signal SSE Support
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		i := 10
		for {
			select {
			case <-r.Context().Done():
				if err := unsubscribeFn(); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

			default:
				b := <-c
				fmt.Fprintf(w, "data: %s\n\n", b)
				i -= 1
				if i <= 0 {
					return
				}
				w.(http.Flusher).Flush()
			}
		}
	}
}

type Notifier interface {
	Notify(b []byte) error
}

type NotificationCenter struct {
	subscribers   map[chan []byte]struct{}
	subscribersMu *sync.Mutex
}

func NewNotificationCenter() *NotificationCenter {
	return &NotificationCenter{
		subscribers:   map[chan []byte]struct{}{},
		subscribersMu: &sync.Mutex{},
	}
}

func (nc *NotificationCenter) Subscribe(c chan []byte) (UnsubscribeFunc, error) {
	nc.subscribersMu.Lock()
	nc.subscribers[c] = struct{}{}
	nc.subscribersMu.Unlock()

	unsubscribeFn := func() error {
		nc.subscribersMu.Lock()
		delete(nc.subscribers, c)
		nc.subscribersMu.Unlock()

		return nil
	}

	return unsubscribeFn, nil
}

func (nc *NotificationCenter) Notify(b []byte) error {
	nc.subscribersMu.Lock()
	defer nc.subscribersMu.Unlock()

	for c := range nc.subscribers {
		select {
		case c <- b:
		default:
		}
	}

	return nil
}