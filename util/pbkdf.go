// pbkdf.go - passphrase based key derivation function
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package util

import (
	"io"

	"github.com/unkaktus/balloon"
	"golang.org/x/crypto/blake2b"
)

var (
	sCost = 1 << 23 // 8 MiB
	tCost = 2
)

func KeystreamReader(passphrase, salt []byte) io.Reader {
	h, err := blake2b.New512(nil)
	if err != nil {
		panic(err)
	}
	secret := balloon.Balloon(h, passphrase, salt, uint64(sCost/h.Size()), uint64(tCost))
	b2xb, err := blake2b.NewXOF(blake2b.OutputLengthUnknown, nil)
	if err != nil {
		panic(err)
	}
	b2xb.Write(secret)
	return b2xb
}
