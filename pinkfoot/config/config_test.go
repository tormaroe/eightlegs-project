package config

import (
	"testing"

	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	doc := []byte(`
	[API]
	Port = 3000
	[Persistance]
	MaxBytes = 1024
	MaxUnreadMessages = 1000`)
	c := Config{}
	toml.Unmarshal(doc, &c)
	assert.Equal(t, 3000, c.API.Port)
	assert.Equal(t, 1024, c.Persistance.MaxBytes)
	assert.Equal(t, 1000, c.Persistance.MaxUnreadMessages)
}

func TestLoad(t *testing.T) {
	c, err := Load("../config.toml")
	if assert.NoError(t, err) {
		assert.Equal(t, 3000, c.API.Port)
		assert.Equal(t, 1024, c.Persistance.MaxBytes)
		assert.Equal(t, 1000, c.Persistance.MaxUnreadMessages)
	}
}
