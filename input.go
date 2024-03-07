package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

var (
	inputChannel = make(chan string)
)

func mindInput() {
	mindGeneric(os.Stdin)
}

func splitAt(substring string) func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	searchBytes := []byte(substring)
	searchLen := len(searchBytes)
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		dataLen := len(data)

		// Return nothing if at end of file and no data passed
		if atEOF && dataLen == 0 {
			return 0, nil, nil
		}

		// Find next separator and return token
		if i := bytes.Index(data, searchBytes); i >= 0 {
			return i + searchLen, data[0:i], nil
		}

		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return dataLen, data, nil
		}

		// Request more data.
		return 0, nil, nil
	}
}

func mindGeneric(r io.Reader) {
	verbose("listening on stdin")
	scanner := bufio.NewScanner(r)
	scanner.Split(splitAt(inputSplit))
	for {
		verbose("waiting on input")
		for scanner.Scan() {
			inputChannel <- scanner.Text()
		}
		verbose("input loop broke")
		if err := scanner.Err(); err != nil {
			workersCancel()
			panic(err)
		}
	}

}
