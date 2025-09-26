package config

import (
	"fmt"
	"os"
	"time"

	"github.com/iden3/go-iden3-auth/v2/pubsignals"
	"github.com/iden3/go-iden3-auth/v2/state"
	"gopkg.in/yaml.v3"
)

// NotificationService is a config for notification service
type NotificationService struct {
	Server                   Server                   `envconfig:"SERVER"`
	Gateway                  Gateway                  `envconfig:"GATEWAY"`
	Redis                    Redis                    `envconfig:"REDIS"`
	Log                      Log                      `envconfig:"LOG"`
	PrivateKey               string                   `envconfig:"PRIVATE_KEY" require:"true"`
	PrivateKeyPath           string                   `envconfig:"PRIVATE_KEY_PATH"`
	ResolversSettingsPath    string                   `envconfig:"RESOLVERS_SETTINGS_PATH" default:"./resolvers.settings.yaml"`
	AuthenticationMiddleware AuthenticationMiddleware `envconfig:"AUTH_MIDDLEWARE"`
}

func (n *NotificationService) GetStateResolvers() (map[string]pubsignals.StateResolver, error) {
	payload := map[string]map[string]struct {
		ContractAddress string `yaml:"contractAddress"`
		NetworkURL      string `yaml:"networkURL"`
	}{}

	resolverFile, err := os.OpenFile(n.ResolversSettingsPath,
		os.O_RDONLY, 0o600)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck // ignore close error
	defer resolverFile.Close()

	if err := yaml.NewDecoder(resolverFile).
		Decode(&payload); err != nil {
		return nil, err
	}

	resolvers := make(map[string]pubsignals.StateResolver)

	for chain, networks := range payload {
		for network, settings := range networks {
			resolver := state.NewETHResolver(settings.NetworkURL, settings.ContractAddress)
			resolvers[fmt.Sprintf("%s:%s", chain, network)] = resolver
		}
	}

	if len(resolvers) == 0 {
		return nil, fmt.Errorf("no resolvers configured in %s", n.ResolversSettingsPath)
	}

	return resolvers, nil
}

// Log config for Log.
type Log struct {
	Env   string `envconfig:"ENV" default:"development"`
	Level string `envconfig:"LEVEL" default:"DEBUG"`
}

// Server config of issuer.
type Server struct {
	Host string `envconfig:"HOST" require:"true"`
	Port int    `envconfig:"PORT" default:"8085"`
}

// Gateway is public gateway config
type Gateway struct {
	Host string `envconfig:"HOST" require:"true"`
}

// Redis config for Redis.
type Redis struct {
	URL                string        `envconfig:"REDIS_URL" required:"true"`
	ExpirationDuration time.Duration `envconfig:"EXPIRATION_DURATION" default:"24h"`
}

// AuthenticationMiddleware is config for auth middleware
type AuthenticationMiddleware struct {
	StateTransitionDelay time.Duration `envconfig:"STATE_TRANSITION_DELAY"`
	ProofGenerationDelay time.Duration `envconfig:"PROOF_GENERATION_DELAY"`
	JWZGenerationDelay   time.Duration `envconfig:"JWZ_GENERATION_DELAY" default:"24h"` // Set 0 to disable jwz rotation
}
