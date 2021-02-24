package main

import (
	"fmt"
	"github.com/schidstorm/sshd-manager/cli"
	"github.com/schidstorm/sshd-manager/config"
	"github.com/schidstorm/sshd-manager/http"
	"github.com/schidstorm/sshd-manager/putter"
)

func main() {
	cli.Run(func(cfg *cli.CliConfig) error {
		config.ParseConfig(cfg.ConfigFile)
		var putterObj putter.Putter
		switch config.Current.Putter {
		case putter.PutterTypeEtcd:
			putterObj = putter.NewEtcd(config.Current.Etcd)
		default:
			return fmt.Errorf("passed invalid putter type. expected '%s', got '%s'", putter.PutterTypeEtcd, config.Current.Putter)
		}
		server := http.NewServer(config.Current.Listen, putterObj)
		err := server.Run(config.Current.Listen)
		if err != nil {
			return err
		}
		return nil
	})
}
