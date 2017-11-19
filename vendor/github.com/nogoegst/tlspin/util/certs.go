// certs.go - generate TLS certificates.
//
// To the extent possible under law, Ivan Markin has waived all copyright
// and related or neighboring rights to tlspin, using the Creative
// Commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.
package tlspinutil

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
)

func RandomSerialNumber() (*big.Int, error) {
	return rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
}

func GenerateEphemeralCert(sk interface{}) ([]byte, error) {
	serialNumber, err := RandomSerialNumber()
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(sk), sk)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %s", err)
	}

	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return pemCert, nil
}

func GeneratePEMKeypair(skstr string) (cert []byte, privkey []byte, err error) {
	var sk []byte
	if skstr == "whateverkey" {
		_, err = io.ReadFull(rand.Reader, sk)
	} else {
		sk, err = DecodeKey(skstr)
	}
	if err != nil {
		return nil, nil, err
	}
	priv, err := privateKey(sk)
	if err != nil {
		return nil, nil, err
	}
	cert, err = GenerateEphemeralCert(priv)
	if err != nil {
		return nil, nil, err
	}
	privPEMBlock, err := MarshalPrivateKeyToPEM(priv)
	if err != nil {
		return nil, nil, err
	}
	privkey = pem.EncodeToMemory(privPEMBlock)
	return cert, privkey, nil
}

func GenerateCertificate(skstr string) (*tls.Certificate, error) {
	cert, privkey, err := GeneratePEMKeypair(skstr)
	if err != nil {
		return nil, err
	}
	tlsCert, err := tls.X509KeyPair(cert, privkey)
	if err != nil {
		return nil, err
	}
	return &tlsCert, nil
}
