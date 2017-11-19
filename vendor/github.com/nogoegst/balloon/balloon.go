// balloon.go - implementation of Balloon memory-hard hashing.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of balloon, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package balloon

import (
	"encoding/binary"
	"hash"
	"math/big"
)

const (
	delta = 3
)

// Instance represents Ballon instance (its internal state).
type Instance struct {
	Buffer    []byte
	LastBlock []byte
	Cnt       uint64
}

// Balloon uses non-memory-hard cryptographic hash function h
// and calculates memory-hard Ballon hash of passphrase with salt.
// sCost is the number of digest-sized blocks in buffer (space cost).
// tCost is the number of rounds (time cost).
func Balloon(h hash.Hash, passphrase, salt []byte, sCost, tCost uint64) []byte {
	b := &Instance{Buffer: make([]byte, sCost*uint64(h.Size()))}
	b.Expand(h, passphrase, salt, sCost)
	b.Mix(h, salt, sCost, tCost)
	return b.LastBlock
}

// BalloonM runs M concurrent Balloon instances and returns
// XOR of their outputs. All other parameters are the same as in Balloon.
func BalloonM(hr func() hash.Hash, passphrase, salt []byte, sCost, tCost uint64, M uint64) []byte {
	out := make([]byte, hr().Size())
	bouts := make(chan []byte)
	for m := uint64(0); m < M; m++ {
		go func(core uint64) {
			binaryM := make([]byte, 8)
			binary.BigEndian.PutUint64(binaryM, core)
			bouts <- Balloon(hr(), passphrase, append(salt, binaryM...), sCost, tCost)
		}(m + 1)
	}
	for m := uint64(0); m < M; m++ {
		for i, v := range <-bouts {
			out[i] ^= v
		}
	}
	return FinalHash(hr(), passphrase, salt, out)
}

// FinalHash hashes output of Balloon function in order to guarantee
// collision and second-preimage resistance (if hash function h provides
// these properties).
func FinalHash(h hash.Hash, passphrase, salt, balloonOut []byte) []byte {
	h.Reset()
	h.Write(passphrase)
	h.Write(salt)
	h.Write(balloonOut)
	return h.Sum(nil)
}

// Expand performs Balloon expansion of (passphrase, salt) using hash function h
// and fills b.Buffer with this output. It panics if size of b.Buffer
// is not sCost*h.Size().
func (b *Instance) Expand(h hash.Hash, passphrase, salt []byte, sCost uint64) {
	blockSize := uint64(h.Size())
	if len(b.Buffer) != int(sCost*blockSize) {
		panic("balloon: internal buffer has wrong length")
	}
	h.Reset()
	binary.Write(h, binary.BigEndian, b.Cnt)
	b.Cnt++
	h.Write(passphrase)
	h.Write(salt)
	b.LastBlock = h.Sum(nil)
	copy(b.Buffer, b.LastBlock)

	for m := uint64(1); m < sCost; m++ {
		h.Reset()
		binary.Write(h, binary.BigEndian, b.Cnt)
		h.Write(b.LastBlock)
		b.LastBlock = h.Sum(nil)
		copy(b.Buffer[b.Cnt*blockSize:], b.LastBlock)
		b.Cnt++
	}
}

// Mix performs Balloon mixing of b.Buffer contents using hash function h and
// salt salt. Mixing parameters are the same as in Balloon: sCost for space cost,
// tCost for number of rounds. It panics if size of b.Buffer is not sCost*h.Size().
func (b *Instance) Mix(h hash.Hash, salt []byte, sCost, tCost uint64) {
	blockSize := uint64(h.Size())
	if len(b.Buffer) != int(sCost*blockSize) {
		panic("balloon: internal buffer has wrong length")
	}
	sCostInt := big.NewInt(int64(sCost))
	otherInt := big.NewInt(0)

	for t := uint64(0); t < tCost; t++ {
		for m := uint64(0); m < sCost; m++ {
			h.Reset()
			binary.Write(h, binary.BigEndian, b.Cnt)
			b.Cnt++
			h.Write(b.LastBlock)
			h.Write(b.Buffer[m*blockSize : (m+1)*blockSize])
			b.LastBlock = h.Sum(nil)
			copy(b.Buffer[m*blockSize:], b.LastBlock)

			for i := uint64(0); i < delta; i++ {
				h.Reset()
				binary.Write(h, binary.BigEndian, b.Cnt)
				b.Cnt++
				h.Write(salt)
				binary.Write(h, binary.BigEndian, t)
				binary.Write(h, binary.BigEndian, m)
				binary.Write(h, binary.BigEndian, i)
				otherInt.SetBytes(h.Sum(nil))
				otherInt.Mod(otherInt, sCostInt)
				other := otherInt.Uint64()
				h.Reset()
				binary.Write(h, binary.BigEndian, b.Cnt)
				b.Cnt++
				h.Write(b.LastBlock)
				h.Write(b.Buffer[other*blockSize : (other+1)*blockSize])
				b.LastBlock = h.Sum(nil)
				copy(b.Buffer[m*blockSize:], b.LastBlock)
			}
		}
	}
}
