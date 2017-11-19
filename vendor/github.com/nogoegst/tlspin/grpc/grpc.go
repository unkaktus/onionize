// grpc.go - tlspin supplementary for gRPC.
//
// To the extent possible under law, Ivan Markin has waived all copyright
// and related or neighboring rights to tlspin, using the Creative
// Commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package tlspingrpc

import (
	"github.com/nogoegst/tlspin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func ServerCredentials(privatekey string) credentials.TransportCredentials {
	tlsConfig, err := tlspin.TLSServerConfig(privatekey)
	if err != nil {
		panic(err)
	}
	return credentials.NewTLS(tlsConfig)
}

func ClientCredentials(publickey string) credentials.TransportCredentials {
	tlsConfig, err := tlspin.TLSClientConfig(publickey)
	if err != nil {
		panic(err)
	}
	return credentials.NewTLS(tlsConfig)
}

func WithPrivateKey(privatekey string) grpc.ServerOption {
	return grpc.Creds(ServerCredentials(privatekey))
}

func WithPublicKey(publickey string) grpc.DialOption {
	return grpc.WithTransportCredentials(ClientCredentials(publickey))
}
