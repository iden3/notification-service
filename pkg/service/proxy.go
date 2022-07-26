package service

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/pkg/errors"
)

type Message struct {
	Content      Content      `json:"content"`
	PushMetadata PushMetadata `json:"metadata"`
}

type PushMetadata struct {
	Devices []EncryptedDeviceMetadata `json:"devices"`
}
type EncryptedDeviceMetadata struct {
	Ciphertext string `json:"ciphertext"` // base64 encoded cipher
	Alg        string `json:"alg"`
}

type ProxyService struct {
	cryptographer *Cryptographer
	notification  *NotificationClient
}

func NewProxy(c *Cryptographer, n *NotificationClient) *ProxyService {
	return &ProxyService{
		cryptographer: c,
		notification:  n,
	}
}

func (ps *ProxyService) Proxy(ctx context.Context, msg *Message) ([]byte, error) {
	var devices []Device
	for _, m := range msg.PushMetadata.Devices {
		d, err := base64.StdEncoding.DecodeString(m.Ciphertext)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		plaintext, err := ps.cryptographer.Decrypt(m.Alg, d)
		if err != nil {
			return nil, errors.Errorf("failed decrypt device info: %s", err)
		}

		var device Device
		err = json.Unmarshal(plaintext, &device)
		if err != nil {
			return nil, errors.Errorf("failed unmarshal device info: %s", err)
		}

		devices = append(devices, device)
	}

	resp, err := ps.notification.Notify(ctx, devices, msg.Content)
	if err != nil {
		return nil, errors.Errorf("failed to notify devices: %s", err)
	}

	return resp, nil
}
