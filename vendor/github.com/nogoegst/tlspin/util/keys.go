// keys.go - operations with keys.
//
// To the extent possible under law, Ivan Markin has waived all copyright
// and related or neighboring rights to tlspin, using the Creative
// Commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package tlspinutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/nogoegst/blake2xb"
	"golang.org/x/crypto/blake2b"
)

const (
	defaultSigningAlg = "P256"
)

func GenerateRawPrivateKey(r io.Reader, signalg string) (sk interface{}, err error) {
	var rsaKeySize int
	if strings.HasPrefix(signalg, "RSA") {
		rsaKeySize, err = strconv.Atoi(strings.TrimPrefix(signalg, "RSA"))
		if err != nil {
			return nil, err
		}
		if rsaKeySize < 2048 {
			return nil, errors.New("RSA key size is too small")
		}
		signalg = "RSA"
	}
	switch signalg {
	case "RSA":
		return rsa.GenerateKey(r, rsaKeySize)
	case "P224":
		return ecdsa.GenerateKey(elliptic.P224(), r)
	case "P256":
		return ecdsa.GenerateKey(elliptic.P256(), r)
	case "P384":
		return ecdsa.GenerateKey(elliptic.P384(), r)
	case "P521":
		return ecdsa.GenerateKey(elliptic.P521(), r)
	}
	return nil, errors.New("unrecognized signing algorithm")
}

func privateKey(sk []byte) (interface{}, error) {
	signalg := defaultSigningAlg
	b2xb, err := blake2xb.New(0)
	if err != nil {
		return nil, err
	}
	b2xb.Write(sk)
	return GenerateRawPrivateKey(b2xb, signalg)
}

func publicKey(sk interface{}) interface{} {
	switch k := sk.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func generateBinaryPrivateKey() ([]byte, error) {
	sk := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, sk)
	if err != nil {
		return nil, err
	}
	return sk, nil
}

func PublicKey(skstr string) (string, error) {
	skdata, err := DecodeKey(skstr)
	if err != nil {
		return "", err
	}
	sk, err := privateKey(skdata)
	if err != nil {
		return "", err
	}
	pk := publicKey(sk)
	der, err := x509.MarshalPKIXPublicKey(pk)
	if err != nil {
		return "", err
	}
	d := blake2b.Sum256(der)
	return EncodeKey(d[:]), nil
}

func GeneratePrivateKey() (string, error) {
	sk, err := generateBinaryPrivateKey()
	if err != nil {
		return "", err
	}
	return EncodeKey(sk), nil
}
