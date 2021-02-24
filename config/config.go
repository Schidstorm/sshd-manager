package config

import (
	"fmt"
	"github.com/schidstorm/sshd-manager/parser"
	"github.com/schidstorm/sshd-manager/putter"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type Config struct {
	Etcd   putter.EtcdConfig `yaml:"etcd"`
	Putter putter.PutterType `yaml:"putter"`
	Listen string            `yaml:"listen"`
}

var Current = &Config{
	Etcd:   putter.EtcdConfig{},
	Putter: putter.PutterTypeEtcd,
	Listen: "127.0.0.1:8080",
}

func ParseConfig(filePath string) *Config {
	buffer, err := ioutil.ReadFile(filePath)
	if err != nil {
		logrus.Warningln(fmt.Sprintf("Config file %s not found. Using default value.", filePath))
	}
	err = parser.ParseYaml(Current, string(buffer))
	if err != nil {
		panic(err)
	}

	return Current
}
