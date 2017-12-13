package config

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type MatrixConfig struct {
	Homeserver  string `yaml:"homeserver"`
	AccessToken string `yaml:"access_token"`
	MXID        string `yaml:"mxid"`
}

type DatabaseConfig struct {
	Path string `yaml:"path,omitempty"`
}

type Feed struct {
	URL          string `yaml:"url"`
	EventType    string `yaml:"event_type"`
	PollInterval int64  `yaml:"poll_interval"`
}

type Config struct {
	Matrix   MatrixConfig   `yaml:"matrix"`
	Feeds    []Feed         `yaml:"feeds"`
	Database DatabaseConfig `yaml:"database"`
}

func Load(filePath string) (cfg Config, err error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		return
	}

	logrus.Info("Config loaded")

	return
}
