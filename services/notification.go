package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/notification-service/log"
	"github.com/pkg/errors"
)

// PushNotification is a structure of message to accept from sender
type PushNotification struct {
	Message      json.RawMessage `json:"message"`
	PushMetadata PushMetadata    `json:"metadata"`
}

func (p *PushNotification) Validate() error {
	if len(p.Message) == 0 {
		return errors.New("message is required")
	}
	if len(p.PushMetadata.Devices) == 0 {
		return errors.New("at least one device is required")
	}
	for _, d := range p.PushMetadata.Devices {
		if d.Ciphertext == "" {
			return errors.New("device ciphertext is required")
		}
		if d.Alg == "" {
			return errors.New("device alg is required")
		}
	}
	return nil
}

// PushMetadata is an array of  encrypted devices info
type PushMetadata struct {
	Devices []EncryptedDeviceMetadata `json:"devices"`
}

// EncryptedDeviceMetadata is an encrypted device info
type EncryptedDeviceMetadata struct {
	Ciphertext string `json:"ciphertext"` // base64 encoded cipher
	Alg        string `json:"alg"`
}

// NotificationResult is a result of msg processing
type NotificationResult struct {
	Device EncryptedDeviceMetadata `json:"device"`
	Status NotificationStatus      `json:"status"`
	Reason string                  `json:"reason"`
}
type cryptoService interface {
	Decrypt(msg []byte) ([]byte, error)
	Encrypt(msg []byte) ([]byte, error)
	Alg() string
}
type cachingService interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, duration time.Duration) error
}

// Notification is a service to notification push notification
type Notification struct {
	notification       *PushClient
	cryptoService      cryptoService
	cachingService     cachingService
	hostURL            string
	expirationDuration time.Duration
}

// NewNotificationService new instance of notification service
func NewNotificationService(n *PushClient, c cryptoService, cs cachingService, host string, expirationDuration time.Duration) *Notification {
	return &Notification{
		notification:       n,
		cryptoService:      c,
		cachingService:     cs,
		hostURL:            host,
		expirationDuration: expirationDuration,
	}
}

// SendNotification sends notification to matrix gateway
func (ns *Notification) SendNotification(ctx context.Context, msg *PushNotification) []NotificationResult {

	msgProcessingResult := make([]NotificationResult, 0)

	decryptedMap := make(map[string]EncryptedDeviceMetadata)

	devices := make([]Device, 0)

	for _, encDeviceInfo := range msg.PushMetadata.Devices {

		device, err := ns.decryptDeviceInfo(encDeviceInfo)
		if err != nil {
			msgProcessingResult = append(msgProcessingResult, NotificationResult{
				Device: encDeviceInfo,
				Status: NotificationStatusFailed,
				Reason: err.Error(),
			})
			continue
		}
		// if device info is valid let's save it's encrypted and decrypted forms
		decryptedMap[device.Pushkey] = encDeviceInfo
		devices = append(devices, device)
	}

	// if there are no valid decrypted device tokens we must return the result immediately
	if len(devices) == 0 {
		return msgProcessingResult
	}

	rejectedTokens, err := ns.notify(ctx, msg, devices)
	if err != nil {
		// return failed for all devices
		for _, device := range devices {
			msgProcessingResult = append(msgProcessingResult, NotificationResult{
				Device: decryptedMap[device.Pushkey],
				Status: NotificationStatusFailed,
				Reason: err.Error(),
			})
		}
		return msgProcessingResult
	}

	// response contains decrypted rejected push tokens. We must return encrypted tokens instead,
	// so sender can exclude encrypted tokens and will not send push again

	for token, enc := range decryptedMap {

		isRejected := contains(rejectedTokens, token)
		if isRejected {
			msgProcessingResult = append(msgProcessingResult, NotificationResult{
				Device: enc,
				Status: NotificationStatusRejected,
				Reason: "Push message could have been rejected by an unstream gateway because they have expired or have never been valid",
			})
			continue
		}
		msgProcessingResult = append(msgProcessingResult, NotificationResult{
			Device: enc,
			Status: NotificationStatusSuccess,
		})
	}

	return msgProcessingResult
}

func (ns *Notification) decryptDeviceInfo(enc EncryptedDeviceMetadata) (Device, error) {
	d, err := base64.StdEncoding.DecodeString(enc.Ciphertext)
	if err != nil {
		return Device{}, errors.New("invalid cipher text format. expected valid base64 encoded string")
	}
	if enc.Alg != ns.cryptoService.Alg() {
		return Device{}, errors.Errorf("service doesn't support %s alg for encrypted device info", enc.Alg)
	}
	plaintext, err := ns.cryptoService.Decrypt(d)
	if err != nil {
		return Device{}, errors.Errorf("service couldn't decrypt the device token")
	}

	var device Device
	err = json.Unmarshal(plaintext, &device)
	if err != nil {
		return Device{}, errors.Errorf("service couldn't process the device token")
	}
	return device, nil

}
func (ns *Notification) notify(ctx context.Context, push *PushNotification, devices []Device) ([]string, error) {

	id := uuid.NewString()
	idToDevices := make(map[string][]Device)
	for _, d := range devices {
		key := buildMessageKey(d.UniqueID, id)
		idToDevices[key] = append(idToDevices[key], d)
	}

	bytesToSave, err := json.Marshal(push.Message)
	if err != nil {
		return nil, errors.New("failed to prepare notification")
	}

	rejects := []string{}
	for saveID, devices := range idToDevices {
		// save a message to a caching service
		err = ns.cachingService.Set(ctx, saveID,
			bytesToSave, ns.expirationDuration)
		if err != nil {
			log.Error(err)
			return nil, errors.New("failed to save device notification")
		}

		u, err := buildResourceURL(ns.hostURL, saveID)
		if err != nil {
			log.Error(err)
			return nil, errors.New("failed to build notification URL")
		}

		contentBody := struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		}{
			ID:  saveID,
			URL: u,
		}

		rawContentBody, err := json.Marshal(contentBody)
		if err != nil {
			log.Error(err)
			return nil, errors.New("failed to notify devices")
		}

		c := Content{
			Body: rawContentBody,
		}
		rejectedTokens, err := ns.notification.SendPush(ctx, devices, c)
		if err != nil {
			log.Error(err)
			return nil, errors.New("failed to notify devices")

		}
		rejects = append(rejects, rejectedTokens...)
	}
	return rejects, nil
}

func buildResourceURL(host, id string) (string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return "", err
	}
	u = u.JoinPath("api", "v1", id)
	return u.String(), nil
}

func buildMessageKey(uniqueID, id string) string {
	if uniqueID == "" {
		return id
	}
	return fmt.Sprintf("%s+%s", uniqueID, id)
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
