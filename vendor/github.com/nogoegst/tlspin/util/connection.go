package tlspinutil

import (
	"crypto/tls"
	"crypto/x509"
	"net"

	"golang.org/x/crypto/blake2b"
)

func InitDialWithDialer(dialer *net.Dialer, network, addr string) (conn net.Conn, keydigest []byte, err error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	c, err := tls.DialWithDialer(dialer, network, addr, tlsConfig)
	if err != nil {
		return nil, nil, err
	}
	connstate := c.ConnectionState()
	chainlen := len(connstate.PeerCertificates)
	if chainlen > 0 {
		peercert := connstate.PeerCertificates[chainlen-1]
		der, _ := x509.MarshalPKIXPublicKey(peercert.PublicKey)
		hash := blake2b.Sum256(der)
		return c, hash[:], nil
	}
	return c, nil, nil
}

func InitDial(network, addr string) (conn net.Conn, keydigest []byte, err error) {
	return InitDialWithDialer(new(net.Dialer), network, addr)
}
