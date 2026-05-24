package config

const defaultListenAddr = "127.0.0.1:8080"

// Config holds runtime settings for the backend.
type Config struct {
	ListenAddr string
}

// DefaultConfig returns the application's default configuration.
func DefaultConfig() Config {
	return Config{ListenAddr: defaultListenAddr}
}

// WithDefaults fills in zero-value fields with defaults.
func (c Config) WithDefaults() Config {
	if c.ListenAddr == "" {
		c.ListenAddr = defaultListenAddr
	}
	return c
}
