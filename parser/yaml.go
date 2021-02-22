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

func SerializeYaml(target interface{}) ([]byte, error) {
	data, err := yaml.Marshal(target)
	if err != nil {
		return nil, err
	}

	return data, nil
}
