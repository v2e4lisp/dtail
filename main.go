package main

import (
	"fmt"
	"os/exec"
)

func main() {
	// output channel
	// child process's stdout/stderr will be send to this channel
	outch := make(chan []byte)

	run(outch, "tail", "-f", "/tmp/main.go")

	for {
		msg := <-outch
		fmt.Print(string(msg))
	}
}

func run(ch chan<- []byte, c string, arg ...string) error {
	cmd := exec.Command(c, arg...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// read from stderr
	go func() {
		for {
			buf := make([]byte, 120)
			stderr.Read(buf)
			ch <- buf
		}
	}()

	// read from stdout
	go func() {
		for {
			buf := make([]byte, 120)
			stdout.Read(buf)
			ch <- buf
		}
	}()

	return nil
}
