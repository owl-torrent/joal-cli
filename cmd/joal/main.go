package main

import (
	"flag"
	"os"
)

func main() {
	var (
		configPath       string
		enableWebUi      bool
		webUiPathPrefix  string
		webUiPort        int
		webUiSecretToken string
	)

	flagSet := flag.NewFlagSet("joal", flag.ContinueOnError)

	currentWorkingDir, _ := os.Getwd()
	flagSet.StringVar(&configPath, "dir", currentWorkingDir, "Joal working directory (the one containing the config file, torrents/, clients/, ...")

	flagSet.BoolVar(&enableWebUi, "webui", false, "Enable the webui")
	flagSet.StringVar(&webUiPathPrefix, "webui.path.prefix", "change-me", "The webui path prefix http://xxx.xxx.xxx.xxx:yyyy/path.prefix/ui")
	flagSet.IntVar(&webUiPort, "webui.port", 9080, "The webui port. For both http and websocket")
	flagSet.StringVar(&webUiSecretToken, "webui.secret.token", "change-it", "The webui secret token")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	//TODO:read config file

	//seedmanager.SeedManagerNew(configPath, seedConfig)

}
