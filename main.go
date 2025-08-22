package main

import (
	"github.com/pyhub/pyhub-docs/cmd"
)

// Version information (set by ldflags during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Set version info for the cmd package
	cmd.Version = version
	cmd.Commit = commit
	cmd.BuildDate = date
	
	cmd.Execute()
}