package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

const path = "/_matrix/push/v1/notify"

type notification struct {
	Devices []Device `json:"devices"`
	Content Content  `json:"content"`
}

// NotificationStatus is a notification status
type NotificationStatus string

const (
	// NotificationStatusSuccess is for pushes that are sent to APNS / FCM
	NotificationStatusSuccess NotificationStatus = "success"
	// NotificationStatusRejected is for pushes that are rejected by APNS / FCM
	NotificationStatusRejected NotificationStatus = "rejected"
	// NotificationStatusFailed is for pushes that were not sent
	NotificationStatusFailed NotificationStatus = "failed"
)

// Device info
type Device struct {
	AppID   string `json:"app_id"`
	Pushkey string `json:"pushkey"`
}

// Content for matrix message
type Content struct {
	Body    json.RawMessage `json:"body"`
	MsgType string          `json:"msgtype"`
}

// PushClient to send push to matrix
type PushClient struct {
	conn *http.Client
	url  string
}

// notificationRes PPG for notify devices.
type notificationRes struct {
	Rejected []string `json:"rejected"`
}

// NewPushClient create PPG client.
func NewPushClient(conn *http.Client, url string) *PushClient {
	return &PushClient{
		conn: conn,
		url:  fmt.Sprintf("%s%s", strings.TrimSuffix(url, "/"), path),
	}
}

// SendPush send push notification in json format to devices.
func (c *PushClient) SendPush(ctx context.Context, listDevices []Device, content Content) ([]string, error) {
	reqData := struct {
		Notification notification `json:"notification"`
	}{
		Notification: notification{
			Devices: listDevices,
			Content: content,
		},
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	notifyRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	respBody, err := c.conn.Do(notifyRequest)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer respBody.Body.Close()

	if respBody.StatusCode != http.StatusOK {
		return nil, errors.New("can't send push notification")
	}
	data, err := io.ReadAll(respBody.Body)
	if err != nil {
		return nil, err
	}

	var pushResult notificationRes
	err = json.Unmarshal(data, &pushResult)
	if err != nil {
		return nil, err
	}

	return pushResult.Rejected, nil
}
