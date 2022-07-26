package config

type Notification struct {
	Server     Server  `envconfig:"SERVER"`
	Gateway    Gateway `envconfig:"GATEWAY"`
	Log        Log     `envconfig:"LOG"`
	PrivateKey string  `envconfig:"PRIVATE_KEY" require:"true"`
}

// Log config for Log.
type Log struct {
	Env   string `envconfig:"ENV" default:"development"`
	Level string `envconfig:"LEVEL" default:"DEBUG"`
}

// Server config of issuer.
type Server struct {
	Port int `envconfig:"PORT" default:"8085"`
}

type Gateway struct {
	Host           string        `envconfig:"HOST" require:"true"`
}
