// blake2xb.go - implementation of BLAKE2Xb.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of blake2xb, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package blake2xb

import (
	"bytes"
	"errors"
	"hash"
	"io"
)

const (
	maxXOFLength = 0xffffffff
)

type BLAKE2xb struct {
	config    *Config      // current config
	rootHash  hash.Hash    // Input hash instance
	h0        []byte       // H0, tree root
	hbuf      bytes.Buffer // Working output buffer
	chainSize uint32       // Number of B2 blocks in XOF chain
}

// Size returns the digest size in bytes.
func (b *BLAKE2xb) Size() int { return int(b.config.Tree.XOFLength) }

// BlockSize returns the algorithm block size in bytes.
func (b *BLAKE2xb) BlockSize() int { return BlockSize }

// Sum returns the calculated checksum.
func (b *BLAKE2xb) Sum(in []byte) []byte {
	hash := make([]byte, b.Size())
	_, err := io.ReadFull(b, hash)
	if err != nil {
		panic(err.Error())
	}
	return append(in, hash[:b.Size()]...)
}

// Write absorbs data on input. It panics if input is written
// after output has been read from the XOF (i.e. Read has been called).
func (x *BLAKE2xb) Write(p []byte) (written int, err error) {
	if x.h0 != nil {
		panic("blake2xb: writing after read")
	}
	return x.rootHash.Write(p)
}

// Read reads output of BLAKE2xb XOF. It returns io.EOF if the end
// of XOF output is reached.
func (x *BLAKE2xb) Read(out []byte) (n int, err error) {
	if x.h0 == nil {
		x.h0 = x.rootHash.Sum(nil)
		setB2Config(x.config)
	}
	dlen := len(out)
	if uint32(dlen) > x.config.Tree.XOFLength {
		return 0, errors.New("blake2xb: destination size is greater than XOF length")
	}
	for x.hbuf.Len() < dlen {
		// Add more blocks
		if x.config.Tree.NodeOffset == x.chainSize {
			x.config.Size = uint8(x.config.Tree.XOFLength % Size)
		}
		b, err := newBlake2b(x.config)
		if err != nil {
			return 0, err
		}
		b.Write(x.h0)
		wn, err := x.hbuf.Write(b.Sum(nil))
		if err != nil {
			return 0, err
		}
		if wn != b.Size() {
			panic("blake2xb: wrong size of written data")
		}
		x.config.Tree.NodeOffset++
	}

	return x.hbuf.Read(out)

}

// Reset resets BLAKE2xb to the initial state.
func (x *BLAKE2xb) Reset() {
	x.rootHash.Reset()
	x.h0 = nil
	x.hbuf.Reset()
	x.config.Size = Size
	x.config.Tree.NodeOffset = 0
}

// NewConfig creates default config c for BLAKE2xb with output length of l.
// If l is 0, maximum output length is used (2^32-1).
func NewConfig(l uint32) (c *Config) {
	if l == 0 {
		l = maxXOFLength
	}
	return &Config{
		Tree: &Tree{XOFLength: l},
	}
}

// NewMAC returns a new hash.Hash computing BLAKE2xb prefix-
// Message Authentication Code of the given size in bytes
// with the given key (up to 64 bytes in length).
func NewMAC(outBytes uint32, key []byte) XOFHash {
	cfg := NewConfig(outBytes)
	cfg.Key = key
	d, err := NewWithConfig(cfg)
	if err != nil {
		panic(err.Error())
	}
	return d
}

func New(l uint32) (XOFHash, error) {
	cfg := NewConfig(l)
	return NewWithConfig(cfg)
}

// NewX creates new BLAKE2xb instance using config c.
func NewWithConfig(c *Config) (XOFHash, error) {
	x := &BLAKE2xb{}
	if c == nil {
		c = NewConfig(maxXOFLength)
	}

	if c.Tree.XOFLength == 0 {
		// Set maximum XOF size if it's zero.
		c.Tree.XOFLength = maxXOFLength
	}

	overrideRootConfig(c)
	if err := verifyConfig(c); err != nil {
		return x, err
	}
	d, err := newBlake2b(c)
	if err != nil {
		return x, err
	}
	x.rootHash = d
	x.chainSize = c.Tree.XOFLength / Size
	x.config = c
	return x, nil
}

func overrideRootConfig(c *Config) {
	// Override size of underlying hash
	c.Size = Size
	// The values below are "as usual".
	// Set them as in reference to match testvectors.
	c.Tree.Fanout = 1
	c.Tree.MaxDepth = 1
	c.Tree.LeafSize = 0
	c.Tree.NodeOffset = 0
	c.Tree.NodeDepth = 0
	c.Tree.InnerHashSize = 0
}

func setB2Config(c *Config) {
	c.Key = nil
	c.Tree.Fanout = 0
	c.Tree.MaxDepth = 0
	c.Tree.LeafSize = Size
	c.Tree.NodeDepth = 0
	c.Tree.InnerHashSize = Size
}
