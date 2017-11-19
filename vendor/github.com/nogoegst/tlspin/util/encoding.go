// encoding.go - encode keys.
//
// To the extent possible under law, Ivan Markin has waived all copyright
// and related or neighboring rights to tlspin, using the Creative
// Commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.
package tlspinutil

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

func MarshalPrivateKeyToPEM(sk interface{}) (*pem.Block, error) {
	switch k := sk.(type) {
	case *rsa.PrivateKey:
		block := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		}
		return block, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal private key: %v", err)
		}
		block := &pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: b,
		}
		return block, nil
	}
	return nil, errors.New("unsupported private key type")
}

func EncodeKey(k []byte) string {
	return base64.StdEncoding.EncodeToString(k)
}

func DecodeKey(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

//LoadKey loads private key from a file in base64-encoded form
func LoadKeyFromFile(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
