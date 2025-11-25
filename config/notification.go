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
	CORS                     CORS                     `envconfig:"CORS"`
	PrivateKey               string                   `envconfig:"PRIVATE_KEY" require:"true"`
	PrivateKeyPath           string                   `envconfig:"PRIVATE_KEY_PATH"`
	ResolversSettingsPath    string                   `envconfig:"RESOLVERS_SETTINGS_PATH" default:"./resolvers.settings.yaml"`
	AuthenticationMiddleware AuthenticationMiddleware `envconfig:"AUTH_MIDDLEWARE"`
	Subscription             Subscription             `envconfig:"SUBSCRIPTION"`
	EnableHTTPPprof          bool                     `envconfig:"ENABLE_HTTP_PPROF" default:"false"`
	SupportedWebAgents       []string                 `envconfig:"SUPPORTED_WEB_AGENTS"`
}

// CORS holds configuration for allowed origins and headers
type CORS struct {
	AllowedOrigins []string `envconfig:"ALLOWED_ORIGINS" default:"https://*"`
	AllowedHeaders []string `envconfig:"ALLOWED_HEADERS" default:"Accept,Authorization,Content-Type,X-CSRF-Token"`
	MaxAge         int      `envconfig:"MAX_AGE" default:"300"`
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
	Host              string        `envconfig:"HOST" require:"true"`
	Port              int           `envconfig:"PORT" default:"8085"`
	ReadHeaderTimeout time.Duration `envconfig:"READ_HEADER_TIMEOUT" default:"10s"`
	ReadTimeout       time.Duration `envconfig:"READ_TIMEOUT" default:"15s"`
	IdleTimeout       time.Duration `envconfig:"IDLE_TIMEOUT" default:"120s"`
	MaxHeaderBytes    int           `envconfig:"MAX_HEADER_BYTES" default:"1048576"`
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
	VerifierDID          string        `envconfig:"VERIFIER_DID" require:"true"`
	StateTransitionDelay time.Duration `envconfig:"STATE_TRANSITION_DELAY"`
	ProofGenerationDelay time.Duration `envconfig:"PROOF_GENERATION_DELAY"`
	JWZGenerationDelay   time.Duration `envconfig:"JWZ_GENERATION_DELAY" default:"24h"` // Set 0 to disable jwz rotation
}

type Subscription struct {
	PingTickerTime       time.Duration `envconfig:"PING_TICKER_TIME" default:"10s"`
	MaxConnectionPerUser int           `envconfig:"MAX_CONNECTION_PER_USER" default:"10"`
	ChannelBufferSize    int           `envconfig:"CHANNEL_BUFFER_SIZE" default:"10"`
}
