package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	// default error message
	errmsg = "Error occured. \n"
	quit   = make(chan string)
	done   = make(chan string)
	out    = make(chan string)
)

func main() {
	args := cli()
	for _, arg := range args {
		t := strings.SplitN(arg, ":", 2)
		if len(t) == 2 {
			panicTail(t[0], remoteCmd(t[0], t[1]))
		} else {
			panicTail(t[0], localCmd(t[0]))
		}
	}

	for {
		select {
		case msg := <-out:
			fmt.Print(msg)
		case err := <-done:
			fmt.Print(err)
			// close all goroutines
			close(quit)
			return
		}
	}
}

func cli() []string {
	flag.Usage = func() {
		fmt.Println("Usage: dtail [server:]file [[server:]file]")
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

func panicTail(tag string, cmd *exec.Cmd) {
	if err := tail(tag, cmd); err != nil {
		panic(err)
	}
}

func tail(tag string, cmd *exec.Cmd) error {
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

	// read from stdout
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				line, err := bufout.ReadBytes('\n')
				if err != nil {
					line, err := buferr.ReadBytes('\n')
					if err != nil {
						done <- tag + errmsg
					} else {
						done <- tag + string(line)
					}
					return
				}
				out <- tag + string(line)
			}
		}
	}()

	return nil
}
