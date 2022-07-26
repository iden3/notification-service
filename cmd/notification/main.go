package main

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"

	"github.com/iden3/signed-notification/pkg/config"
	"github.com/iden3/signed-notification/pkg/log"
	"github.com/iden3/signed-notification/pkg/rest"
	"github.com/iden3/signed-notification/pkg/rest/handlers"
	"github.com/iden3/signed-notification/pkg/service"
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
		log.Fatal("failed decode pem foramt")
	}
	privK, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err != nil {
		log.Fatal("failed decode private key:", err)
	}

	cryptographer, err := service.NewCryptographerService(privK)
	if err != nil {
		log.Fatal("failed init cryptographer:", err)
	}

	c := &http.Client{Transport: &retryablehttp.RoundTripper{}}
	notificationClient := service.NewClient(c, cfg.Gateway.Host)
	proxyService := service.NewProxy(cryptographer, notificationClient)

	h := rest.NewHandlers(handlers.NewProxyHandler(proxyService), handlers.NewKeyHandler(cryptographer))
	r := h.Routes()
	server := rest.NewServer(r)
	err = server.Run(cfg.Server.Port)
	if errors.Cause(err) == http.ErrServerClosed {
		log.Info("HTTP server stopped")
	} else if err != nil {
		log.Errorf("HTTP server error: %v", err)
	}
}
