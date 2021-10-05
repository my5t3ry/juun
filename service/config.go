package main

type Config struct {
	AutoSaveInteralSeconds uint
	SerchResults           int
	EnableVowpalWabbit     bool
	LogLevel               string
}

func NewConfig() *Config {
	return &Config{
		AutoSaveInteralSeconds: 300,
		SerchResults:           30,
		EnableVowpalWabbit:     true,
		LogLevel:               "info",
	}
}
