package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type telegram struct {
	Key    string `yaml:"key"`
	Debug  bool   `yaml:"debug"`
	Enable bool   `yaml:"enable"`
}

type Database struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Config struct {
	Database *Database `yaml:"database"`
	Telegram telegram  `yaml:"telegram"`
}

func NewConfig(f string) (*Config, error) {
	rawConf, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("cannot open a Config: %s\n", err.Error())
	}
	conf := Config{}
	err = yaml.Unmarshal(rawConf, &conf)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshall a Config: %s\n", err.Error())
	}
	return &conf, nil
}
