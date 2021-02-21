package parser

import (
	"gopkg.in/yaml.v2"
)

func ParseYaml(target interface{}, yamlString string) error {
	err := yaml.Unmarshal([]byte(yamlString), target)
	if err != nil {
		return err
	}

	return nil
}