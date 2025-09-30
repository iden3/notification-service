package services

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/pkg/errors"
)

const (
	// list of supported algorithms.
	rsaAlg = "RSA-OAEP-512"
)

// Crypto is a service to encrypt and decrypt device data with a presetup key
type Crypto struct {
	publicKey  crypto.PublicKey
	privateKey crypto.Decrypter
	alg        string
}

// NewCryptoService creates new instance of crypto
func NewCryptoService(pk crypto.PrivateKey) (*Crypto, error) {
	var alg string
	switch pk.(type) {
	case *rsa.PrivateKey:
		alg = rsaAlg
	default:
		return nil, errors.Errorf("alg %s in not supported by service", alg)
	}
	k := pk.(crypto.Decrypter)
	return &Crypto{
		publicKey:  k.Public(),
		privateKey: k,
		alg:        alg,
	}, nil
}

// MarshalPubKeyToPem converts public key to pem format
func (cr *Crypto) MarshalPubKeyToPem() ([]byte, error) {
	raw, err := x509.MarshalPKIXPublicKey(cr.publicKey)
	if err != nil {
		return nil, err
	}
	b := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: raw,
	}
	w := bytes.NewBuffer([]byte{})
	err = pem.Encode(w, b)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// Decrypt encrypts given byte array with setup key
func (cr *Crypto) Decrypt(msg []byte) ([]byte, error) {

	var (
		plaintext []byte
		err       error
	)
	switch cr.alg {
	case rsaAlg:
		plaintext, err = cr.privateKey.Decrypt(rand.Reader, msg, &rsa.OAEPOptions{
			// TODO(illia-korotia): for more flexibility, we can specify a hash function
			// in request. like we set algorithm
			Hash:  crypto.SHA512,
			Label: nil,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
	default:
		return nil, fmt.Errorf("decryption is not supported for alg: %s", cr.alg)
	}
	return plaintext, nil
}

// Encrypt encrypts given byte array with setup key
func (cr *Crypto) Encrypt(msg []byte) ([]byte, error) {

	switch cr.alg {
	case rsaAlg:
		encryptedBytes, err := rsa.EncryptOAEP(
			sha512.New(),
			rand.Reader,
			cr.publicKey.(*rsa.PublicKey),
			msg,
			nil)
		if err != nil {
			return nil, err
		}
		return encryptedBytes, nil
	default:
		return nil, fmt.Errorf("encryption is not supported for alg: %s", cr.alg)
	}
}

// Alg returns current alg of crypto service
func (cr *Crypto) Alg() string {
	return cr.alg
}
