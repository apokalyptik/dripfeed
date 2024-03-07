package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	command    []string
	processes  = 1
	window     = 10 * time.Second
	execOut    = ""
	inputSplit = "\n"
	debug      bool
)

func init() {
	if v := os.Getenv("DRIPFEED_PROCESSES"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			processes = i
		}
	}

	if v := os.Getenv("DRIPFEED_WINDOW"); v != "" {
		if t, err := time.ParseDuration(v); err == nil {
			window = t
		}
	}

	if v := os.Getenv("DRIPFEED_EXEC"); v != "" {
		execOut = v
	}

	if v := os.Getenv("DRIPFEED_INPUT_SPLIT"); v != "" {
		inputSplit = v
	}

	flag.IntVar(&processes, "processes", processes, "The number of commands to warm up and have ready [DRIPFEED_COMMAND]")
	flag.DurationVar(&window, "window", window, "The window of time to give input to for each instance of the command [DRIPFEED_WINDOW]")
	flag.BoolVar(&debug, "verbose", debug, "Be Verbose in logging output [DRIPFEED_VERBOSE]")
	flag.StringVar(&execOut, "exec", execOut, "Instead of sending data to stdout, execute a command and send data to its stdin [DRIPFEED_EXEC]")
	flag.StringVar(&inputSplit, "input-split", inputSplit, "Input should be split by this string instead of newlines. Use an empry string for null. [DRIPFEED_INPUT_SPLIT]")

	flag.Usage = func() {
		fmt.Fprint(
			flag.CommandLine.Output(),
			strings.Join(
				[]string{
					filepath.Base(os.Args[0]) + " [flags] [commmand [args [...]]]",
					"",
					filepath.Base(os.Args[0]) + " is a tool for dividing up a long running or infinite stream",
					"of data from stdin by slices of time and sending that data to the",
					"stdin of other processes.",
					"",
					"For example lets say you want to count the number of lines coming",
					"into a log file every 30 seconds in perpetuity. The problem with",
					"simply running `tail -F access_log | wc -l` is that the wc process",
					"will wait infinitely long before yielding results. But " + filepath.Base(os.Args[0]) + " can",
					"get what you need. by running the following:",
					"",
					"`tail -F access_log | " + filepath.Base(os.Args[0]) + " -window 30s wc -l`",
					"",
					"From which we will get output like this:",
					"",
					"	13114",
					"	11971",
					"	12334",
					"",
					"What " + filepath.Base(os.Args[0]) + " does is execute 3, by default, instances of wc (you",
					"can see this in a process list if you check in another terminal.)",
					"It then begins taking lines that it gets from its stdin and",
					"copying those to the first instance of wc. At the end of the first",
					"window of time " + filepath.Base(os.Args[0]) + " will do a number of things all at the same",
					"time.",
					"",
					"	1. It will begin sending new lines to the second wc process",
					"	2. It will spin up a 4th wc process to be ready and waiting",
					"	   for its input",
					"	3. It will close stdin to the first wc process, and wait",
					"	   for it to exit",
					"",
					"In this way " + filepath.Base(os.Args[0]) + " will keep cycling through wc processes for all",
					"eternity, processing an infinite input divided by windows of time.",
					"",
					"It is important that the command being fed data to its stdin detect when",
					"its stdin has been closed and exits gracefully on its own. Otherwise we",
					"will end up with an ever growing number of running programs that never go",
					"away",
					"",
					"This is the default mode of operation for " + filepath.Base(os.Args[0]) + ". It is assumed",
					"that you will do something with the stdout stream from " + filepath.Base(os.Args[0]) + "",
					"itself. This mode is good for summary logging, or responding to",
					"changes in a stream of data programmatically.",
					"",
					"Sometimes, however, you need to process the output generated in",
					"discreet chunks and not as a stream. To accomplish this we have the",
					"-exec flag. For example if we run:",
					"",
					"	`tail -F access_log | " + filepath.Base(os.Args[0]) + " -exec 'md5sum' -window 10s wc -l`",
					"",
					"Then we might get the following output:",
					"",
					"	690b5ca7d1a758f171ab412e7890807c  -",
					"	82522fa50b09cc9339024f2c3cc22eda  -",
					"	d8eb074366fb9f7533038e198cc69013  -",
					"",
					"This time, for each instance of wc that " + filepath.Base(os.Args[0]) + " starts it also",
					"starts an instance of md5sum. The difference here is that the",
					"stdout of wc is being sent to the stdin of md5sum. So instead of",
					"getting the output of wc we get the output of md5sum.",
					"",
					"This is, of course a contrived example and the expected usage pattern",
					"is to replace md5sum with a program or script which will do something",
					"based on data fed to it via stdin, such as a reporting script or",
					"program",
					"",
					"It is important that the command being run by -exec detects that stdin",
					"has been closed and self terminates. Otherwise you will end up with an",
					"infinitely growing number of processes hanging around for no reason.",
					"",
					fmt.Sprintf("Usage of %s:\n", filepath.Base(os.Args[0])),
				},
				"\n",
			),
		)
		flag.PrintDefaults()
		os.Exit(0)
	}
}

func main() {
	flag.Parse()

	if processes < 1 {
		panic("Worker count must be greater than or equal to 1")
	}

	command = flag.Args()
	if len(command) < 1 {
		panic("A command must be provided")
	}

	if inputSplit == "" {
		inputSplit = "\000"
	}

	ticker = time.Tick(window)
	workers = make(chan *workerProcess, processes)

	go mindQueue()
	go mindWorkers()
	mindInput()
}
