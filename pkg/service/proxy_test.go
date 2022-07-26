package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

const mockJson = `{"notification":{"devices":[{"app_id":"local.id","pushkey":"cyoy-rV4Yls7HG3P5vDn5j:APA91bGSRgegCsNBXIwTeHvWCgMExvmVINl3r8RYZFG0MxtKdw_zIiJIft1m0V0etDOGIPDYOVNU6NuZ_S9yELw2veT_9ZOZsYXoY_3bdDT38c-eb6oAoj0Lq3rgY5YZmWc0t6JWFgYJ"}],"content":{"body":{"id":"123321"},"msgtype":"type/json"}}}`

func signalMock(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, mockJson, string(data))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"rejected":[]}`))
	}))
}

func TestProxy(t *testing.T) {

	keypair, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)
	cripter, err := NewCryptographerService(keypair)
	require.NotNil(t, cripter)
	require.NoError(t, err)

	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type: "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(keypair),
		},
	)
	log.Println("private ket in pem:", string(pemdata))

	signal := signalMock(t)
	defer signal.Close()
	notificationClient := NewClient(http.DefaultClient, signal.URL)

	// mock signal with http_test.
	proxy := NewProxy(cripter, notificationClient)

	device := Device{
		AppID:   "local.id",
		Pushkey: "cyoy-rV4Yls7HG3P5vDn5j:APA91bGSRgegCsNBXIwTeHvWCgMExvmVINl3r8RYZFG0MxtKdw_zIiJIft1m0V0etDOGIPDYOVNU6NuZ_S9yELw2veT_9ZOZsYXoY_3bdDT38c-eb6oAoj0Lq3rgY5YZmWc0t6JWFgYJ",
	}
	encodedDevice, err := json.Marshal(device)
	require.NoError(t, err)

	// get public key in pem format
	pemPubKey, err := proxy.cryptographer.MarshalToPemPublicKey()
	require.NoError(t, err)
	pubk, _ := pem.Decode(pemPubKey)
	require.NotNil(t, pubk)
	pk, err := x509.ParsePKIXPublicKey(pubk.Bytes)
	require.NoError(t,err)

	ciphertext, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, pk.(*rsa.PublicKey), encodedDevice, nil)
	require.NoError(t, err)

	log.Println("example:", base64.StdEncoding.EncodeToString(ciphertext))

	msg := &Message{
		Content: Content{
			Body:    []byte(`{"id": "123321"}`),
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

	_, err = proxy.Proxy(context.Background(), msg)
	require.NoError(t, err)
}
