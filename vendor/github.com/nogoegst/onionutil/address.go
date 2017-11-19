// address.go - commonly used functions for onion addresses
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionutil, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package onionutil

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"errors"
	"io"
	"reflect"

	"github.com/nogoegst/onionutil/pkcs1"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
)

// Generate private key for onion service using rand as the entropy source.
// Recognized versions are "2", "3", "current", "best".
func GenerateOnionKey(rand io.Reader, version string) (crypto.PrivateKey, error) {
	switch version {
	case "2", "current":
		return GenerateOnionKeyV2(rand)
	case "3", "best":
		return GenerateOnionKeyV3(rand)
	default:
		return nil, errors.New("Unrecognized version string for onion address")
	}
}

// OnionAddress returns onion address corresponding to public/private key pk.
func OnionAddress(pk crypto.PublicKey) (string, error) {
	switch pk := pk.(type) {
	case *rsa.PublicKey:
		return OnionAddressV2(pk)
	case *rsa.PrivateKey:
		return OnionAddress(pk.Public().(*rsa.PublicKey))
	case ed25519.PublicKey:
		return OnionAddressV3(pk)
	case ed25519.PrivateKey:
		return OnionAddressV3(pk.Public().(ed25519.PublicKey))
	default:
		return "", errors.New("Unrecognized type of public key")
	}
}

// Check whether onion address is a valid one.
func OnionAddressIsValid(onionAddress string) bool {
	v2v := OnionAddressIsValidV2(onionAddress)
	v3v := OnionAddressIsValidV3(onionAddress)
	return v2v || v3v
}

// v2 onion addresses
var (
	OnionAddressLengthV2 = 10
)

// OnionAddress returns the Tor Onion Service address corresponding to a given
// rsa.PublicKey.
func OnionAddressV2(pk *rsa.PublicKey) (onionAddress string, err error) {
	permID, err := CalcPermanentID(pk)
	if err != nil {
		return onionAddress, err
	}
	onionAddress = Base32Encode(permID)
	return onionAddress, err
}

// Generate v2 onion service key (RSA-1024) using rand as the entropy source.
func GenerateOnionKeyV2(rand io.Reader) (crypto.PrivateKey, error) {
	sk, err := rsa.GenerateKey(rand, 1024)
	if err != nil {
		return nil, err
	}
	return sk, nil
}

// Check whether onion address is a valid v2 one.
func OnionAddressIsValidV2(onionAddress string) bool {
	oa, err := Base32Decode(onionAddress)
	if err != nil {
		return false
	}
	if len(oa) != OnionAddressLengthV2 {
		return false
	}
	return true
}

// Calculate hash (SHA1) of DER-encoded RSA public key pk.
func RSAPubkeyHash(pk *rsa.PublicKey) (derHash []byte, err error) {
	der, err := pkcs1.EncodePublicKeyDER(pk)
	if err != nil {
		return
	}
	derHash = Hash(der)
	return derHash, err
}

// Calculate permanent ID from RSA public key
func CalcPermanentID(pk *rsa.PublicKey) (permId []byte, err error) {
	derHash, err := RSAPubkeyHash(pk)
	if err != nil {
		return
	}
	permId = derHash[:10]
	return
}

// v3 onion addresses
var (
	OnionAddressChecksumLengthV3     = 2
	OnionAddressVersionFieldV3       = []byte{0x03}
	OnionAddressVersionFieldLengthV3 = 1
	OnionAddressLengthV3             = ed25519.PublicKeySize +
		OnionAddressVersionFieldLengthV3 +
		OnionAddressChecksumLengthV3
	OnionChecksumPrefix = []byte(".onion checksum")
)

// Calculate onion address v3 from public key pk.
func OnionAddressV3(pk ed25519.PublicKey) (onionAddress string, err error) {
	chksum := OnionAddressChecksumV3([]byte(pk))
	oab := make([]byte, 0, OnionAddressLengthV3)
	oa := bytes.NewBuffer(oab)
	oa.Write([]byte(pk))
	oa.Write(chksum)
	oa.Write(OnionAddressVersionFieldV3)
	onionAddress = Base32Encode(oa.Bytes())
	return onionAddress, err
}

// Check whether onion address is a valid v3 one.
func OnionAddressIsValidV3(onionAddress string) bool {
	_, err := OnionAddressPublicKeyV3(onionAddress)
	return err == nil
}

// Extract Ed25519 public key from the onion address.
func OnionAddressPublicKeyV3(onionAddress string) (ed25519.PublicKey, error) {
	oa, err := Base32Decode(onionAddress)
	if err != nil {
		return nil, errors.New("Error while base32 decoding onion address")
	}
	if len(oa) != OnionAddressLengthV3 {
		return nil, errors.New("Wrong onion address length")
	}
	oab := bytes.NewBuffer(oa)
	pk := oab.Next(ed25519.PublicKeySize)
	chksum := oab.Next(OnionAddressChecksumLengthV3)
	ver := oab.Next(OnionAddressVersionFieldLengthV3)
	if !reflect.DeepEqual(ver, OnionAddressVersionFieldV3) {
		return nil, errors.New("Invalid onion address version value")
	}
	if !reflect.DeepEqual(chksum, OnionAddressChecksumV3(pk)) {
		return nil, errors.New("Invalid onion address checksum")
	}
	return ed25519.PublicKey(pk), nil
}

// Generate v3 onion address key (Ed25519) using rand as the entropy source
func GenerateOnionKeyV3(rand io.Reader) (crypto.PrivateKey, error) {
	_, sk, err := ed25519.GenerateKey(rand)
	return sk, err
}

// Calculate onion address checksum (v3) from byte-encoded Ed25519 key
func OnionAddressChecksumV3(pk []byte) []byte {
	h := sha3.New256()
	h.Write(OnionChecksumPrefix)
	h.Write([]byte(pk))
	h.Write(OnionAddressVersionFieldV3)
	return h.Sum(nil)[:2]
}
