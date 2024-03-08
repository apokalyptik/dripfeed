# dripfeed

dripfeed is a tool for dividing up a long running or infinite stream
of data to stdin by slices of time and sending that data to the
stdin of other processes.

# simple usage

Lets say you want to count the number of lines coming
into a log file every 30 seconds in perpetuity. The problem with
simply running `tail -F access_log | wc -l` is that the wc process
will wait infinitely long before yielding results. But dripfeed can
get what you need. by running the following:

```
tail -F access_log | dripfeed -window 30s wc -l
```
From which we will get output like this:

```
	13114
	11971
	12334
```

What dripfeed does is execute 3, by default, instances of wc (you
can see this in a process list if you check in another terminal.)
It then begins taking lines that it gets from its stdin and
copying those to the first instance of wc. At the end of the first
window of time dripfeed will do a number of things all at the same
time.

1. It will begin sending new lines to the second wc process
2. It will spin up a 4th wc process to be ready and waiting for its input
3. It will close stdin to the first wc process, and wait for it to exit

In this way dripfeed will keep cycling through wc processes for all
eternity, processing an infinite input divided by windows of time.

It is _important_ that the command being fed data to its stdin detect when
its stdin has been closed and exits gracefully on its own. Otherwise we
will end up with an ever growing number of running programs that never go
away

This is the default mode of operation for dripfeed. It is assumed
that you will do something with the stdout stream from dripfeed
itself. This mode is good for summary logging, or responding to
changes in a stream of data programmatically.

# advanced usage

Sometimes, however, you need to process the output generated in
discreet chunks and not as a stream. To accomplish this we have the
-exec flag. For example if we run:

```
tail -F access_log | dripfeed -exec 'md5sum' -window 10s wc -l
```

Then we might get the following output:

```
	690b5ca7d1a758f171ab412e7890807c  -
	82522fa50b09cc9339024f2c3cc22eda  -
	d8eb074366fb9f7533038e198cc69013  -
```

This time, for each instance of wc that dripfeed starts it also
starts an instance of md5sum. The difference here is that the 
data will flow as follows:

`--> [dripfeed stdin] --> [wc -l] --> [md5sum] --> [dripfeed stdout]`

This model of execution is helpful for times when you have a need
to preprocess the output before using it in a tool over which you
have little or no control. For example: filtering and sanitizing the 
discreet chunks of a logfile before running them through a log
analysis tool such as [andle-grinder](https://github.com/rcoh/angle-grinder)
which you would invoke like so

```
tail -F access_log | dripfeed -exec 'agrind {query}' -window 10s filterscript
```

It is _important_ that the command being run by -exec detects that stdin
has been closed and self terminates. Otherwise you will end up with an
infinitely growing number of processes hanging around for no reason.

# invoking

```
Usage of dripfeed:
  -exec string
    	Instead of sending data to stdout, execute a command and send data to its stdin [DRIPFEED_EXEC]
  -input-split string
    	Input should be split by this string instead of newlines. Use an empry string for null. [DRIPFEED_INPUT_SPLIT] (default "\n")
  -processes int
    	The number of commands to warm up and have ready [DRIPFEED_COMMAND] (default 1)
  -verbose
    	Be Verbose in logging output [DRIPFEED_VERBOSE]
  -window duration
    	The window of time to give input to for each instance of the command [DRIPFEED_WINDOW] (default 10s)
```
