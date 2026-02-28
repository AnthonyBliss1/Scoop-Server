package main

import (
	"flag"

	"github.com/anthonybliss1/Scoop-Server/cmd"
)

func main() {
	port := flag.Int("port", 2767, "http server port number")
	deploy := flag.Bool("deploy", false, "deploy systemd or launchd service")

	flag.Parse()

	switch *deploy {
	case true:
		cmd.StartDeploy(port)

	default:
		cmd.StartServer(port)
	}
}
