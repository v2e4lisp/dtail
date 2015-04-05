package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// output channel
	// child process's stdout/stderr will be send to this channel
	outch := make(chan string)

	args := cli()
	for _, arg := range args {
		t := strings.SplitN(arg, ":", 2)
		if len(t) == 2 {
			panicTail(t[0], outch, remoteCmd(t[0], t[1]))
		} else {
			panicTail(t[0], outch, localCmd(t[0]))
		}
	}

	for {
		msg := <-outch
		fmt.Print(msg)
	}
}

func cli() []string {
	flag.Usage = func() {
		fmt.Println("Usage: tailf [server:]file [[server:]file]")
		os.Exit(1)
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
	}

	return flag.Args()
}

func remoteCmd(s string, f string) *exec.Cmd {
	cmd := "tail -f " + f
	return exec.Command("ssh", s, cmd)
}

func localCmd(f string) *exec.Cmd {
	return exec.Command("tail", "-f", f)
}

func panicTail(tag string, ch chan<- string, cmd *exec.Cmd) {
	if err := tail(tag, ch, cmd); err != nil {
		panic(err)
	}
}

func tail(tag string, ch chan<- string, cmd *exec.Cmd) error {
	if tag != "" {
		tag = "[" + strings.TrimSpace(tag) + "] "
	}

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

	bufout, buferr := bufio.NewReader(stdout), bufio.NewReader(stderr)

	// read from stderr
	go func() {
		for {
			line, _ := buferr.ReadBytes('\n')
			ch <- tag + string(line)
		}
	}()

	// read from stdout
	go func() {
		for {
			line, _ := bufout.ReadBytes('\n')
			ch <- tag + string(line)
		}
	}()

	return nil
}
