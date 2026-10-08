// This file exposes the bounded process and preview-server commands to site builds.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "typography:", err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	if len(arguments) == 2 && arguments[0] == "process" {
		return processDirectory(arguments[1])
	}
	if len(arguments) >= 1 && arguments[0] == "serve" {
		return serveHugo(arguments[1:])
	}
	return fmt.Errorf("usage: typography process <directory> | typography serve [--bind <address>] [--baseURL <url>]")
}
