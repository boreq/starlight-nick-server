// Package config holds the configuration struct.
package config

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ServeAddress string
	DatabasePath string
}

// Default returns the default config.
func Default() *Config {
	conf := &Config{
		ServeAddress: "127.0.0.1:8118",
		DatabasePath: "path/to/database.bolt",
	}
	return conf
}

// Load loads the specified config file.
func Load(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	if err := json.Unmarshal(content, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
