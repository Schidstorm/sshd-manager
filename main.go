package main

import (
	"github.com/schidstorm/sshd-manager/cli"
	"github.com/schidstorm/sshd-manager/config"
	"github.com/schidstorm/sshd-manager/http"
)

func main() {
	cli.Run(func(cfg *cli.CliConfig) error {
		config.ParseConfig(cfg.ConfigFile)
		err := http.Run(config.Current.Listen)
		if err != nil {
			return err
		}
		return nil
	})
}
