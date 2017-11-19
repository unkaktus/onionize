// pbkdf.go - passphrase based key derivation function
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionutil, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package onionutil

import (
	"encoding/hex"
	"io"

	"github.com/nogoegst/balloon"
	"github.com/nogoegst/blake2xb"
	"github.com/dchest/blake2b"
)

var (
	sCost          = 1 << 23 // 8 MiB
	tCost          = 2
	saltBalloon, _ = hex.DecodeString("8e8a1b3347da2672fa404eaa7276dee3")
	saltXOF, _     = hex.DecodeString("313e86e72658f5c7c3ad6e1c3d397062")
)

func KeystreamReader(passphrase []byte, person []byte) (io.Reader, error) {
	h := blake2b.New512()
	secret := balloon.Balloon(h, passphrase, saltBalloon, uint64(sCost/h.Size()), uint64(tCost))

	b2xbConfig := blake2xb.NewConfig(0)
	b2xbConfig.Salt = saltXOF[:16]
	b2xbConfig.Person = person[:16]
	b2xb, err := blake2xb.NewWithConfig(b2xbConfig)
	if err != nil {
		return nil, err
	}
	b2xb.Write(secret)

	return b2xb, nil
}
