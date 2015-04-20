package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	errmsg = "Error occured. \n"
	quit   = make(chan bool)
	// command line args & options
	useTag bool
	fs     []string
)

// TODO: flag usage

func main() {
	setup()
	for _, arg := range fs {
		t := strings.SplitN(arg, ":", 2)
		if len(t) == 2 {
			panicTail(t[0], remoteCmd(t[0], t[1]))
		} else {
			panicTail(t[0], localCmd(t[0]))
		}
	}

	<-quit
	os.Exit(1)
}

func setup() {
	flag.Usage = func() {
		fmt.Printf("dtail [OPTION] [SERVER:]FILE...\n")
		flag.PrintDefaults()
	}
	flag.BoolVar(&useTag, "tag", false, "output with tag")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("No file is found")
		os.Exit(1)
	}
	fs = flag.Args()
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

func puts(tag, line string) {
	if useTag {
		fmt.Print("[" + tag + "] " + line)
	} else {
		fmt.Print(line)
	}
}

func tail(tag string, cmd *exec.Cmd) error {
	tag = strings.TrimSpace(tag)

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

	go func() {
		for {
			line, err := bufout.ReadBytes('\n')
			if err != nil {
				puts(tag, errmsg)
				break
			}
			puts(tag, string(line))
		}
	}()

	go func() {
		line, err := buferr.ReadBytes('\n')
		for err != io.EOF {
			puts(tag, string(line))
			line, err = buferr.ReadBytes('\n')
		}
		quit <- true
	}()

	return nil
}
