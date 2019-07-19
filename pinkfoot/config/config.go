package config

import (
	"io/ioutil"

	toml "github.com/pelletier/go-toml"
)

// Config represents all configuration options
type Config struct {
	API             API
	Persistance     Persistance
	Acknowledgement Acknowledgement
}

// API holds configuration properties for the API
type API struct {
	Port int
}

// Persistance holds configuration properties related to message persistance
type Persistance struct {
	DataFile          string
	NextMessageFile   string
	MaxBytes          int
	MaxUnreadMessages int
}

// Acknowledgement holds configuration properties related to acknowledging message receipt
type Acknowledgement struct {
	SecondsBeforeReInsert int
}

// Load will read and parse TOML configuration from file
func Load(filename string) (Config, error) {
	config := Config{}
	bytes, err := ioutil.ReadFile(filename)
	if err == nil {
		err = toml.Unmarshal(bytes, &config)
	}
	return config, err
}
