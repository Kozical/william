// +build windows

package app

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type (
	// Config is the structure used to unmarshal the config.yaml file
	Config struct {
		PSPath         string `yaml:"ps_path"`
		PSOpts         string `yaml:"ps_opts"`
		ScriptsPath    string `yaml:"scripts_path"`
		MaxConnections int    `yaml:"max_connections"`
		BindAddr       string `yaml:"bind_addr"`
		BindPort       string `yaml:"bind_port"`
		KeyPath        string `yaml:"key_path"`
		CrtPath        string `yaml:"crt_path"`
	}
)

// Init parses command line flags and unmarshal the configuration yaml pointed to by 'path'
func Init(path string) (*Config, error) {

	data, err := readConfig(path)
	if err != nil {
		return nil, err
	}

	config, err := unmarshalConfig(data)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func readConfig(path string) ([]byte, error) {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("File does not exist %s\n", path)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file %s -> %s\n", path, err)
	}

	return data, nil
}

func unmarshalConfig(data []byte) (*Config, error) {
	config := new(Config)

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal configuration -> %s\n", err)
	}

	return config, nil
}
