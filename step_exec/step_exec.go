// step_exec executes commands step by step.
// e.g.,
//   % echo icanhazip.com | step_exec -cmd 'curl -x 127.0.0.1:2020'
//
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
	cmd  string
	step int
)

func init() {
	flag.StringVar(&cmd, "cmd", "/bin/echo", "absolute path to command")
	flag.IntVar(&step, "step", 32, "steps")
}

// Command takes args as an array.
func Command(name string, args []string) *exec.Cmd {
	c := &exec.Cmd{}
	if len(args) > 1 {
		c.Path = name
		c.Args = args
	} else {
		c.Path = name
	}
	return c
}

func join(tasks *int, done chan error) {
	fmt.Printf("have %d tasks\n", *tasks)
	for i := 0; i < *tasks; i++ {
		select {
		case err := <-done:
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v", err)
			}
		}
	}
	*tasks = 0
}

func main() {
	flag.Parse()
	scanner := bufio.NewScanner(os.Stdin)
	done := make(chan error)
	// Space-separated?
	execCmd := strings.Split(cmd, " ")
	path := execCmd[0]
	fmt.Printf("bin: %s, args: %v\n", path, execCmd)
	var current int
	for scanner.Scan() {
		item := scanner.Text()
		current++
		args := make([]string, len(execCmd))
		copy(args, execCmd)
		command := Command(path, args)
		command.Args = append(command.Args, item)
		err := command.Start()
		if err != nil {
			panic(err)
		}
		fmt.Printf("+WIP: %s %5v\n", string(cmd), command.Args)
		go func() {
			done <- command.Wait()
		}()
		if current > step {
			// Wait til all jobs are done.
			join(&current, done)
		}
	}
	fmt.Println("done")
}
