package main

import (
	"log"
	"os"
	"time"
)

var (
	ticker <-chan time.Time
)

func mindQueue() {
	verbose("Waiting on first worker")
	var worker = <-workers
	verbose("New Worker %d", worker.Process.Pid)
	for {
		select {
		case <-ticker:
			go dismissWorker(worker)
			worker = <-workers
			break
		case buf := <-inputChannel:
			verbose("Sending input to worker %d", worker.Process.Pid)
			worker.stdin.Write(append([]byte(buf), '\n'))
			break
		}
	}
}

func dismissWorker(worker *workerProcess) {
	verbose("Dismissing worker %d", worker.Process.Pid)
	if err := worker.stdin.Close(); err != nil {
		log.Printf("Error closing worker %d stdin: %s", worker.Process.Pid, err.Error())
		if err := worker.Process.Kill(); err != nil {
			log.Printf("Error killing worker %d: %s", worker.Process.Pid, err.Error())
		}
	}

	if err := worker.Wait(); err != nil {
		log.Printf("Error waiting on worker %d: %s", worker.Process.Pid, err.Error())
		if err := worker.Process.Kill(); err != nil {
			log.Printf("Error killing worker %d: %s", worker.Process.Pid, err.Error())
		}
	}

	if worker.stdoutWorker != nil {
		if err := worker.stdoutWorker.Wait(); err != nil && err != os.ErrProcessDone {
			log.Printf("Error waiting on stdout worker %d: %s", worker.stdoutWorker.Process.Pid, err.Error())
			if err := worker.stdoutWorker.Process.Kill(); err != nil && err != os.ErrProcessDone {
				log.Printf("Error killing stdout worker %d: %s", worker.stdoutWorker.Process.Pid, err.Error())
			}
		}
	}

	verbose("Worker %d has been dismissed", worker.Process.Pid)
}
