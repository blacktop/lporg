package utils

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

var (
	normalPadding = cli.Default.Padding
)

// Indent indents apex log line to supplied level
func Indent(f func(s string), level int) func(string) {
	return func(s string) {
		cli.Default.Padding = normalPadding * level
		f(s)
		cli.Default.Padding = normalPadding
	}
}

// StringInSlice finds string in array
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// AppendIfMissing with append a string to a array if it doesn't already contain that string
func AppendIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

func checkError(err error) {
	if err != nil {
		log.WithError(err).Fatal("failed")
	}
}

// RunCommand runs cmd on file
func RunCommand(ctx context.Context, cmd string, args ...string) (string, error) {

	var c *exec.Cmd

	if ctx != nil {
		c = exec.CommandContext(ctx, cmd, args...)
	} else {
		c = exec.Command(cmd, args...)
	}

	output, err := c.Output()
	if err != nil {
		return string(output), err
	}

	// check for exec context timeout
	if ctx != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command %s timed out", cmd)
		}
	}

	return string(output), nil
}
