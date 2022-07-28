package main

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/go-redis/redis/v8"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/iden3/notification-service/config"
	"github.com/iden3/notification-service/log"
	"github.com/iden3/notification-service/rest"
	"github.com/iden3/notification-service/rest/handlers"
	"github.com/iden3/notification-service/services"
	"github.com/pkg/errors"
	"net/http"
)

func main() {
	cfg, err := config.ParseNotificationConfig()
	if err != nil {
		log.Fatal("failed parse notification config:", err)
	}
	log.SetEnv(cfg.Log.Env)
	// set log level from config
	log.SetLevelStr(cfg.Log.Level)

	b, _ := pem.Decode([]byte(cfg.PrivateKey))
	if b == nil {
		log.Fatal("failed decode pem format")
	}
	privK, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err != nil {
		log.Fatal("failed decode private key:", err)
	}

	cryptoService, err := services.NewCryptoService(privK)
	if err != nil {
		log.Fatal("failed init crypto service:", err)
	}

	c := &http.Client{Transport: &retryablehttp.RoundTripper{}}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.URL,
		Password: cfg.Redis.Password,
	})

	cachingService := services.NewRedisCacheService(redisClient)
	notificationClient := services.NewPushClient(c, cfg.Gateway.Host)
	notificationService := services.NewNotificationService(notificationClient, cryptoService, cachingService, cfg.Server.Host)

	h := rest.NewHandlers(handlers.NewPushNotificationHandler(notificationService, cachingService), handlers.NewKeyHandler(cryptoService))
	r := h.Routes()
	server := rest.NewServer(r)
	err = server.Run(cfg.Server.Port)
	if errors.Cause(err) == http.ErrServerClosed {
		log.Info("HTTP server stopped")
	} else if err != nil {
		log.Errorf("HTTP server error: %v", err)
	}
}
