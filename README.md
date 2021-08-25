# Log Service Problem

## Problem Assumptions
We do not assume any length on the log files, nor do we assume line length. 
Each event is a single line. I didn't have a Linux Machine available to test, but 
this has been tested on Mac. 

I assumed "query for a specific file in /var/log" meant to retrieve all file contents.
You can also get the last n events of a valid specified file. Finally, you can apply
basic keyword filter. This will return all (or just the last n) log events that have the 
specified non-trivial keyword in it. 

Because of time constraints, this project works with 1 client issuing a REST request
to the server. 

## Structure of Project & Basic Design

Given that the server might have to send over a whole large log line, the 
server listens for a connection from a client. Once it receives the REST request, 
it sends the log lines over as a server side event. Once the server is ready to 
send over relevant log lines

I implemented 1 REST GET Endpoint where the user can specify the file & other
query parameters. After all, the user is just retrieving some information.

The cmd/main.go file contains HTTP Server code which listens for HTTP Requests.
The main goroutine will have a channel that listens for any new lines read.
Another goroutine is responsible for executing the reading logic and sending 
the lines/events read chunk by chunk

## Testing
You can run: go run ./cmd/main.go if you have go set up. 

You can open a client and send a GET request to localhost port 8001: 
curl -v "http://localhost:8001/query?filename=wifi.log&events=7&includefilter=00000000"

This example request here tries to get 7 of the most recent events with the string
00000000 from the wifi.log file.

file_processor_test.go and log_scanner_test.go also provide basic unit
tests. I would have added more unit tests to test more edge cases if 
time permitted. The unit test in file_processor_test.go uses a small 
test log file (test.log).

## Further Optimizations


