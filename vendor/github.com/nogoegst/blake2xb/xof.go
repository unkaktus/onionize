// xor.go - definition of XOF interface.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of blake2xb, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package blake2xb

import (
	"hash"
	"io"
)

type XOFHash interface {
	hash.Hash
	io.Reader
}
