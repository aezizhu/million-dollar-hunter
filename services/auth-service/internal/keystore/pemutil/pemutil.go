package pemutil

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

func EncodeRSAPrivateKeyToPEM(key *rsa.PrivateKey) []byte {
	b := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: b}
	return pem.EncodeToMemory(block)
}

func EncodeRSAPublicKeyToPEM(key *rsa.PublicKey) []byte {
	b := x509.MarshalPKCS1PublicKey(key)
	block := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: b}
	return pem.EncodeToMemory(block)
}

func ParseRSAPrivateKeyFromPEM(p []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(p)
	if block == nil {
		return nil, errors.New("no pem block")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func ParseRSAPublicKeyFromPEM(p []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(p)
	if block == nil {
		return nil, errors.New("no pem block")
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}
