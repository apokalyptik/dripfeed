package main

import "log"

func verbose(s string, input ...interface{}) {
	if !debug {
		return
	}
	log.Printf(s, input...)
}
