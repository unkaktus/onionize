// keyfile.go - operations with onion keyfiles.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionutil, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package onionutil

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

func LoadPrivateKeyFile(filename string) (crypto.PrivateKey, crypto.PublicKey, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}
	block, rest := pem.Decode(fileContent)
	if len(rest) == len(fileContent) {
		return nil, nil, fmt.Errorf("No vailid PEM blocks found")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		sk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, nil, err
		}
		return sk, sk.Public(), err
	default:
		return nil, nil, fmt.Errorf("Unrecognized type of PEM block")
	}
}
