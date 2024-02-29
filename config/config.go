package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	RPCURL          string `yaml:"rpc_url"`
	PrivateKey      string `yaml:"private_key"`
	ContractAddress string `yaml:"contract_address"`
}

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
