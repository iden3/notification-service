package config

import "github.com/kelseyhightower/envconfig"

// ParseNotificationConfig parses env variables in to Notification config.
func ParseNotificationConfig() (*Notification, error) {
	var cfg Notification
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
