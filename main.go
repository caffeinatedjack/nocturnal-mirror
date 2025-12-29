package main

import "gitlab.com/caffeinatedjack/nocturnal/cmd"

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	cmd.Version = Version
	cmd.BuildTime = BuildTime
	cmd.Execute()
}
