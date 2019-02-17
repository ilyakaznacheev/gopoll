package poll

import (
	"errors"
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

// ConfigDB database configuration
type ConfigDB struct {
	Address  string  `yaml:"address"`
	Name     string  `yaml:"db-name"`
	Username *string `yaml:"username"`
	Password *string `yaml:"password"`
}

// ConfigAdmin administrator configuration
type ConfigAdmin struct {
	Name string `yaml:"name"`
	Pass string `yaml:"password"`
}

// ConfigPoll default poll creation settings
type ConfigPoll struct {
	Questions int `yaml:"questions"`
	Answers   int `yaml:"answers"`
}

// Config is a configuration file structure
type Config struct {
	Admin ConfigAdmin `yaml:"admin"`
	DB    ConfigDB    `yaml:"database"`
	Poll  ConfigPoll  `yaml:"poll"`
}

func loadConfig(path string) (*Config, error) {
	confFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.New("config error: " + err.Error())
	}

	conf := &Config{}
	err = yaml.Unmarshal(confFile, conf)
	if err != nil {
		return nil, errors.New("config error: " + err.Error())
	}
	return conf, nil
}
