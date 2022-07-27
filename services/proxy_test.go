package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const mockPushKey = `cyoy-rV4Yls7HG3P5vDn5j:APA91bGSRgegCsNBXIwTeHvWCgMExvmVINl3r8RYZFG0MxtKdw_zIiJIft1m0V0etDOGIPDYOVNU6NuZ_S9yELw2veT_9ZOZsYXoY_3bdDT38c-eb6oAoj0Lq3rgY5YZmWc0t6JWFgYJ`

func signalRejectedMock(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Contains(t, string(data), mockPushKey)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"rejected":["cyoy-rV4Yls7HG3P5vDn5j:APA91bGSRgegCsNBXIwTeHvWCgMExvmVINl3r8RYZFG0MxtKdw_zIiJIft1m0V0etDOGIPDYOVNU6NuZ_S9yELw2veT_9ZOZsYXoY_3bdDT38c-eb6oAoj0Lq3rgY5YZmWc0t6JWFgYJ"]}`))
	}))
}
func signalSuccessMock(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Contains(t, string(data), mockPushKey)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"rejected":[]}`))
	}))
}

type RedisMock struct {
}

func (r RedisMock) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	return nil
}

func (r RedisMock) Get(ctx context.Context, key string) (interface{}, error) {
	return nil, nil
}
func TestProxy(t *testing.T) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)
	cs, err := NewCryptoService(privateKey)
	require.NotNil(t, cs)
	require.NoError(t, err)

	signal := signalSuccessMock(t)
	defer signal.Close()
	notificationClient := NewNotificationClient(http.DefaultClient, signal.URL)
	redisMock := RedisMock{}

	// mock signal with http_test.
	proxy := NewProxyService(notificationClient, cs, redisMock, "host")

	device := Device{
		AppID:   "local.id",
		Pushkey: mockPushKey,
	}
	encodedDevice, err := json.Marshal(device)
	require.NoError(t, err)

	ciphertext, err := proxy.cryptoService.Encrypt(encodedDevice)
	require.NoError(t, err)

	msg := &Message{
		Content: Content{
			Body:    []byte(`{"my_cat": "123321"}`),
			MsgType: "type/json",
		},
		PushMetadata: PushMetadata{
			Devices: []EncryptedDeviceMetadata{
				{
					Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
					Alg:        "rsa",
				},
			},
		},
	}

	res, err := proxy.SendNotification(context.Background(), msg)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, NotificationStatusSuccess, res[0].Status)

}
func TestProxyRejected(t *testing.T) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)
	cs, err := NewCryptoService(privateKey)
	require.NotNil(t, cs)
	require.NoError(t, err)

	signal := signalRejectedMock(t)
	defer signal.Close()
	notificationClient := NewNotificationClient(http.DefaultClient, signal.URL)
	redisMock := RedisMock{}

	// mock signal with http_test.
	proxy := NewProxyService(notificationClient, cs, redisMock, "host")

	device := Device{
		AppID:   "local.id",
		Pushkey: mockPushKey,
	}
	encodedDevice, err := json.Marshal(device)
	require.NoError(t, err)

	ciphertext, err := proxy.cryptoService.Encrypt(encodedDevice)
	require.NoError(t, err)

	msg := &Message{
		Content: Content{
			Body:    []byte(`{"my_cat": "123321"}`),
			MsgType: "type/json",
		},
		PushMetadata: PushMetadata{
			Devices: []EncryptedDeviceMetadata{
				{
					Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
					Alg:        "rsa",
				},
			},
		},
	}

	res, err := proxy.SendNotification(context.Background(), msg)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, NotificationStatusRejected, res[0].Status)

}
func TestProxyFailed(t *testing.T) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)
	cs, err := NewCryptoService(privateKey)
	require.NotNil(t, cs)
	require.NoError(t, err)

	signal := signalRejectedMock(t)
	defer signal.Close()
	notificationClient := NewNotificationClient(http.DefaultClient, signal.URL)
	redisMock := RedisMock{}

	// mock signal with http_test.
	proxy := NewProxyService(notificationClient, cs, redisMock, "host")

	msg := &Message{
		Content: Content{
			Body:    []byte(`{"my_cat": "123321"}`),
			MsgType: "type/json",
		},
		PushMetadata: PushMetadata{
			Devices: []EncryptedDeviceMetadata{
				{
					Ciphertext: base64.StdEncoding.EncodeToString([]byte("mockedInvalidCipherText")),
					Alg:        "rsa",
				},
			},
		},
	}

	res, err := proxy.SendNotification(context.Background(), msg)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, NotificationStatusFailed, res[0].Status)
	require.Equal(t, "service couldn't decrypt the device token", res[0].Reason)

}
