package main

import "gitlab.com/caffeinatedjack/nocturnal/cmd"

// Version and BuildTime are set at build time via ldflags
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	cmd.Version = Version
	cmd.BuildTime = BuildTime
	cmd.Execute()
}
