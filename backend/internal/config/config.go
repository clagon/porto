package config

const defaultListenAddr = "127.0.0.1:8080"

// Config holds runtime settings for the backend.
type Config struct {
	ListenAddr   string `json:"listen_addr"`
	AutoDiscover bool   `json:"auto_discover"`
}

// DefaultConfig returns the application's default configuration.
func DefaultConfig() Config {
	return Config{
		ListenAddr:   defaultListenAddr,
		AutoDiscover: true,
	}
}

// WithDefaults fills in zero-value fields with defaults.
func (c Config) WithDefaults() Config {
	defaults := DefaultConfig()
	if c.ListenAddr == "" {
		c.ListenAddr = defaults.ListenAddr
	}
	if !c.AutoDiscover {
		c.AutoDiscover = defaults.AutoDiscover
	}
	return c
}
