package config

import "time"

type Notification struct {
	Server     Server
	Gateway    Gateway
	Log        Log    `envconfig:"LOG"`
	PrivateKey string `envconfig:"PRIVATE_KEY" require:"true"`
}

// Log config for Log.
type Log struct {
	Env   string `envconfig:"ENV" default:"development"`
	Level string `envconfig:"LEVEL" default:"DEBUG"`
}

// Server config of issuer.
type Server struct {
	Port int `envconfig:"PORT" default:"8001"`
	// Host is used together with port e.g. http://localhost:8001
	Host            string        `envconfig:"HOST"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
}

type Gateway struct {
	Host           string        `envconfig:"HOST" require:"true"`
	RequestTimeOut time.Duration `envconfig:"REQUEST_TIME_OUT" default:"5s"`
}
