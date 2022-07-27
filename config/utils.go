package config

import "github.com/kelseyhightower/envconfig"

// ParseNotificationConfig parses env variables in to NotificationService config.
func ParseNotificationConfig() (*NotificationService, error) {
	var cfg NotificationService
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
