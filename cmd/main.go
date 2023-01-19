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
	"os"
)

func main() {
	cfg, err := config.ParseNotificationConfig()
	if err != nil {
		log.Fatal("failed parse notification config:", err)
	}
	log.SetEnv(cfg.Log.Env)
	// set log level from config
	log.SetLevelStr(cfg.Log.Level)

	var b *pem.Block
	b, _ = pem.Decode([]byte(cfg.PrivateKey))

	if cfg.PrivateKeyPath != "" && b == nil {
		fileContent, err := os.ReadFile(cfg.PrivateKeyPath)
		if err != nil {
			log.Fatal("failed open file with pem content")
		}
		b, _ = pem.Decode(fileContent)
	}
	if b == nil {
		log.Fatal("failed decode pem format")
	}

	var privKey interface{}
	privKey, err = x509.ParsePKCS8PrivateKey(b.Bytes)
	if err != nil {
		privKey, err = x509.ParsePKCS1PrivateKey(b.Bytes)
		if err != nil {
			log.Fatal("failed decode private key:", err)
		}
	}

	cryptoService, err := services.NewCryptoService(privKey)
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
	notificationService := services.NewNotificationService(notificationClient, cryptoService, cachingService, cfg.Server.Address())

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
