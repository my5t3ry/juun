package config

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path"

	. "../common"
)

type Config struct {
	AutoSaveIntervalSeconds uint
	SearchResults           int
	ResultLineLength        int
	EnableVowpalWabbit      bool
	LogLevel                string
}

func NewConfig() *Config {
	return &Config{
		AutoSaveIntervalSeconds: 120,
		SearchResults:           30,
		ResultLineLength:        100,
		EnableVowpalWabbit:      true,
		LogLevel:                "info",
	}
}

func GetConfig() *Config {
	configFile := path.Join(GetHome(), ".juun.config")
	config := NewConfig()
	dat, err := ioutil.ReadFile(configFile)
	if err == nil {
		err = json.Unmarshal(dat, config)
		if err != nil {
			config = NewConfig()
		}
		log.Debugf("config[%s]: %s", configFile, PrettyPrint(config))
	} else {
		log.Warnf("missing config file %s, using default: %s", configFile, PrettyPrint(config))
	}
	return config
}
