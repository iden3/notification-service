package service

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/pkg/errors"
)

const (
	// list of supported algorithms.
	rsaAlg = "rsa"
)

type Cryptographer struct {
	publicKey  crypto.PublicKey
	privateKey crypto.Decrypter
	alg        string
}

func NewCryptographerService(pk crypto.PrivateKey) (*Cryptographer, error) {
	var alg string
	switch pk.(type) {
	case *rsa.PrivateKey:
		alg = rsaAlg
	default:
		return nil, errors.New("unknown cryptography algorithm")
	}
	k := pk.(crypto.Decrypter)
	return &Cryptographer{
		publicKey:  k.Public(),
		privateKey: k,
		alg:        alg,
	}, nil
}

func (cr *Cryptographer) MarshalToPemPublicKey() ([]byte, error) {
	raw, err := x509.MarshalPKIXPublicKey(cr.publicKey)
	if err != nil {
		return nil, err
	}
	b := &pem.Block{
		Type:    "PUBLIC KEY",
		Bytes: raw,
	}
	w := bytes.NewBuffer([]byte{})
	err = pem.Encode(w, b)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (cr *Cryptographer) Decrypt(alg string, msg []byte) ([]byte, error) {
	if cr.alg != alg {
		return nil, fmt.Errorf("'%s' alg doesn't support", alg)
	}
	var (
		plaintext []byte
		err       error
	)
	switch cr.alg {
	case rsaAlg:
		plaintext, err = cr.privateKey.Decrypt(rand.Reader, msg, &rsa.OAEPOptions{
			Hash:  crypto.SHA512,
			Label: nil,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return plaintext, nil
}
