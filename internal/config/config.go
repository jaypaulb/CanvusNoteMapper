package config

import "sync"

// Config holds application configuration
type Config struct {
	MCSServer string
	APIKey    string
}

var (
	currentConfig = &Config{}
	mu            sync.RWMutex
)

// GetConfig returns the current config
func GetConfig() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return &Config{
		MCSServer: currentConfig.MCSServer,
		APIKey:    currentConfig.APIKey,
	}
}

// SetConfig updates the current config
func SetConfig(cfg *Config) {
	mu.Lock()
	defer mu.Unlock()
	currentConfig.MCSServer = cfg.MCSServer
	currentConfig.APIKey = cfg.APIKey
}

// LoadConfig loads configuration (stub)
func LoadConfig() *Config {
	// TODO: Load from env or file
	return GetConfig()
}
