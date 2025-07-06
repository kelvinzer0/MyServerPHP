package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config struct holds all configuration for our application
type Config struct {
	Server ServerConfig `yaml:"server"`
	Auth   Auth         `yaml:"auth"`
	ServersConfigPath string `yaml:"servers_config_path"`
	
}

// ACMEConfig struct holds ACME (Let's Encrypt) configuration


// ServerConfig struct holds server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

// Auth struct holds authentication configuration
type Auth struct {
	Username string `yaml:"username"`
	PasswordHash string `yaml:"password_hash"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create a new Config struct
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

