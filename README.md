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




## Testing

## Further Optimizations


