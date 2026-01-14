package main

import "github.com/lucas-stellet/wk/cmd"

// version is set via ldflags during build.
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
