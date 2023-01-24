package config

import "fmt"

// NotificationService is a config for notification service
type NotificationService struct {
	Server         Server  `envconfig:"SERVER"`
	Gateway        Gateway `envconfig:"GATEWAY"`
	Redis          Redis   `envconfig:"REDIS"`
	Log            Log     `envconfig:"LOG"`
	PrivateKey     string  `envconfig:"PRIVATE_KEY" require:"true"`
	PrivateKeyPath string  `envconfig:"PRIVATE_KEY_PATH"`
}

// Log config for Log.
type Log struct {
	Env   string `envconfig:"ENV" default:"development"`
	Level string `envconfig:"LEVEL" default:"DEBUG"`
}

// Server config of issuer.
type Server struct {
	Protocol     string `envconfig:"PROTOCOL" default:"https"`
	Host         string `envconfig:"HOST" require:"true"`
	Port         int    `envconfig:"PORT" default:"8085"`
	ExternalPort int    `envconfig:"EXTERNAL_PORT" default:"80"`
}

// Address return agent address.
func (s Server) Address() string {
	return fmt.Sprintf("%s://%s:%d", s.Protocol, s.Host, s.ExternalPort)
}

// Gateway is public gateway config
type Gateway struct {
	Host string `envconfig:"HOST" require:"true"`
}

// Redis config for Redis.
type Redis struct {
	URL      string `envconfig:"REDIS_URL" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
}
