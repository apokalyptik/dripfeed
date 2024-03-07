package main

import (
	"context"
	"io"
	"os"
	"os/exec"
)

type workerProcess struct {
	*exec.Cmd
	stdin        io.WriteCloser
	stdoutWorker *exec.Cmd
}

var (
	workers       chan *workerProcess
	workersCancel context.CancelFunc
	workersCtx    context.Context
	fauxStdout    io.Writer
)

func mindWorkers() {
	workersCtx, workersCancel = context.WithCancel(context.Background())
	fauxStdout = io.MultiWriter(os.Stdout)
	for {
		var newCommand *exec.Cmd
		var newStdoutWorker *exec.Cmd

		if len(command) < 2 {
			newCommand = exec.CommandContext(workersCtx, command[0])
		} else {
			newCommand = exec.CommandContext(workersCtx, command[0], command[1:]...)
		}

		f, err := newCommand.StdinPipe()
		if err != nil {
			workersCancel()
			panic(err)
		}

		if execOut != "" {
			workerStdOut, err := newCommand.StdoutPipe()
			if err != nil {
				workersCancel()
				panic(err)
			}
			newStdoutWorker = exec.CommandContext(workersCtx, "/bin/bash", "-c", execOut)
			newStdoutWorker.Stdin = workerStdOut
			newStdoutWorker.Stdout = fauxStdout
			newStdoutWorker.Stderr = os.Stderr
			if err := newStdoutWorker.Start(); err != nil {
				workersCancel()
				panic(err)
			}
		} else {
			newCommand.Stdout = fauxStdout
		}
		newCommand.Stderr = os.Stderr

		if err := newCommand.Start(); err != nil {
			workersCancel()
			panic(err)
		}

		verbose("Prepared worker %d", newCommand.Process.Pid)

		workers <- &workerProcess{
			Cmd:          newCommand,
			stdin:        f,
			stdoutWorker: newStdoutWorker,
		}
	}
}
