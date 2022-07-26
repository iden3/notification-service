package service

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

type Device struct {
	AppID   string `json:"app_id"`
	Pushkey string `json:"pushkey"`
}

type Content struct {
	Body    json.RawMessage `json:"body"`
	MsgType string `json:"msgtype"`
}

// NotificationClient PPG for notify devices.
type NotificationClient struct {
	conn *http.Client
	url  string
}

// NewClient create PPG client.
func NewClient(conn *http.Client, url string) *NotificationClient {
	return &NotificationClient{
		conn: conn,
		url:  fmt.Sprintf("%s%s", strings.TrimSuffix(url, "/"), path),
	}
}

// Notify send notification in json format to devices.
func (c *NotificationClient) Notify(ctx context.Context, listDevices []Device, content Content) ([]byte, error) {
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

	data, err := io.ReadAll(respBody.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
