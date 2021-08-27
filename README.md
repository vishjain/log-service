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
to the server. I used Golang for this project.

## Structure of Project & Basic Design

Given that the server might have to send over a whole large log file, the 
server listens for a connection from a client. Once it receives the REST request, 
it sends the log lines over as SSE (server side events). Once the server is ready to 
send over relevant log lines, it will emit an event. I felt polling would create
too much overhead on the server if the log file as large.

I implemented 1 REST GET Endpoint where the user can specify the file & other
query parameters. After all, the user is just retrieving some information.

The cmd/main.go file contains HTTP Server code which listens for HTTP Requests.
The main goroutine will have a channel that listens for any new lines read.
Another goroutine is responsible for executing the reading logic and sending 
the lines/events read chunk by chunk.

The log scanner file has the underlying implementation to read a file line by line.
It tracks how much has been read and the file pointer position. The log scanner file
reads the file in larger chunks (configured to 4096 * 16 bytes, more on this # later). 
Then you do a basic character search for the new line character to get the last line. 
The larger reads were done to reduce the # of disk accesses/system calls when processing 
larger files.

The file_manager.go and file_processor.go code files use the log scanner
to read the file and send a block of lines back to the main goroutine.
The main goroutine then writes/flushes those lines to the client & goes back to
listening for any more lines/errors. The side goroutine reading from the 
log file sends a configurable block of lines to the main goroutine 
(specified as maxLinesToRetrieve in the file manager). This was done for a 
few reasons:
1) For larger workloads, I felt there could be a performance hit if the main 
goroutine writes just one line, flushes it, and has to listen for another line 
for larger files. You can look at performance experiments for more details.
2) I originally wanted to render query results in a custom-built UI. 
For the client, working with a few lines would be easier than 
listening for a new line/event repeatedly.
3) Probably need websockets/dual-communication for the stretch challenge 
(master node, multiple nodes). This mechanism makes it easier to transition 
into that. You have finer grained control over how many events you want. 
Master node will probably have to message back and forth with the other
worker nodes. 
4) Browsers can have limits/restrictions and this mechanism makes it easier
to deal with such challenges. 

More in the next section as to why maxLinesToRetrieve is set to 320 lines.

## Testing/Performance
You can run: go run ./cmd/main.go if you have go set up. 

You can open a client and send a GET request to localhost port 8001: 
curl -v "http://localhost:8001/query?filename=wifi.log&events=7&includefilter=00000000"

This example request here tries to get 7 of the most recent events with the string
00000000 from the wifi.log file.

file_processor_test.go and log_scanner_test.go also provide basic unit
tests. I would have added more unit tests to test more edge cases if 
time permitted. The unit test in file_processor_test.go uses a small 
test log file (test.log).

To optimize performance for larger workloads, I timed a curl request for a 2GB
randomly generated character file. To try to control for any MacOS caching effects, I would
create a new 2GB randomly generated file every time I pulled a measurement. First,
I tuned how much 1 file read call should attempt to read in each call (blockSizeRead). Inspired 
by the page size of 4096 bytes in linux, I tried 4096 bytes, 4096 * 4 bytes, 4096 * 16 bytes.
I found that the pipeline took 1 minute, 24,45 seconds, then 23.8 seconds. Increasing
the read granularity further didn't affect performance too much - so I kept the parameter
at 4096 * 16 bytes. I imagine pre-fetching/caching at the OS level helps performance as you
read a file backwards, but I'd need to research this more.

Next, I tweaked the maxLinesToRetrieve parameter. Specifically, how many events/lines should
be written & flushed to the client before listening for more events. The above experiments
were done with a 100 lines. I Tried 200 and got 14.55 seconds and 320 and got 13.70 seconds.
More set of lines didn't seem to improve performance too much. I'm assuming that batch
writing & flushing decreased the roundtrip of waiting on the channel for more unflushed 
events. If each read is 4096 * 16 bytes and each line is averaging to 200 bytes, then roughly 300
lines would be read into the buffer in 1 read call. 


## Further Optimizations
I would have implemented more optimizations if I had more time. We could
implement all sorts of caching mechanisms to improve query speed. Here a
few:
1) Cache recent 100-200 lines of a recently queried file.
2) Pre-process files and look for words that occur frequently. Have a 
map that maps the word to which line it occurs in.
