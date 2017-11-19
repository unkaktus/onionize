// tlspin.go - reduce TLS to keypinning.
//
// To the extent possible under law, Ivan Markin has waived all copyright
// and related or neighboring rights to tlspin, using the Creative
// Commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package tlspin

import (
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"

	util "github.com/nogoegst/tlspin/util"
	"golang.org/x/crypto/blake2b"
)

var commonTLSConfig = &tls.Config{
	MinVersion:               tls.VersionTLS12,
	CurvePreferences:         []tls.CurveID{tls.X25519},
	PreferServerCipherSuites: true,
	CipherSuites: []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	},
}

func TLSServerConfig(privatekey string) (*tls.Config, error) {
	tlsCert, err := util.GenerateCertificate(privatekey)
	if err != nil {
		return nil, err
	}
	config := commonTLSConfig.Clone()
	config.Certificates = []tls.Certificate{*tlsCert}
	return config, nil
}

func Listen(network, addr, privatekey string) (net.Listener, error) {
	tlsConfig, err := TLSServerConfig(privatekey)
	if err != nil {
		return nil, err
	}
	return tls.Listen(network, addr, tlsConfig)
}

func verifyPeerCert(rawCerts [][]byte, publickey string) error {
	if publickey == "whateverkey" {
		return nil
	}
	pk, err := util.DecodeKey(publickey)
	if err != nil {
		return err
	}
	if len(rawCerts) == 0 {
		return errors.New("no certificates")
	}
	rcert := rawCerts[len(rawCerts)-1]
	certs, err := x509.ParseCertificates(rcert)
	if err != nil {
		return err
	}
	der, err := x509.MarshalPKIXPublicKey(certs[0].PublicKey)
	if err != nil {
		return err
	}
	hash := blake2b.Sum256(der)
	if subtle.ConstantTimeCompare(hash[:], pk) != 1 {
		return errors.New("invalid key")
	}
	return nil
}

func TLSClientConfig(publickey string) (*tls.Config, error) {
	config := commonTLSConfig.Clone()
	config.InsecureSkipVerify = true
	config.VerifyPeerCertificate = func(rawCerts [][]byte, vc [][]*x509.Certificate) error {
		return verifyPeerCert(rawCerts, publickey)
	}
	return config, nil
}

func DialWithDialer(dialer *net.Dialer, network, addr, publickey string) (net.Conn, error) {
	tlsConfig, err := TLSClientConfig(publickey)
	if err != nil {
		return nil, err
	}
	c, err := tls.DialWithDialer(dialer, network, addr, tlsConfig)
	return c, err
}

func Dial(network, addr, publickey string) (net.Conn, error) {
	return DialWithDialer(new(net.Dialer), network, addr, publickey)
}

func Keyfile(filename string) string {
	sk, _ := util.LoadKeyFromFile(filename)
	return sk
}
