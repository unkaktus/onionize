// blake2xb_test.go - tests for the implementation of BLAKE2Xb.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of blake2xb, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package blake2xb

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"
)

type testVector struct {
	In  string
	Key string
	Out string
}

func TestXOF(t *testing.T) {
	f, err := os.Open("testvectors/blake2xb.json")
	if err != nil {
		t.Errorf("Unable to open testvectors file: %v", err)
	}
	dec := json.NewDecoder(f)
	_, err = dec.Token()
	if err != nil {
		t.Error(err)
	}
	for dec.More() {
		var v testVector
		err := dec.Decode(&v)
		if err != nil {
			t.Error(err)
		}
		in, err := hex.DecodeString(v.In)
		if err != nil {
			t.Error(err)
		}
		out, err := hex.DecodeString(v.Out)
		if err != nil {
			t.Error(err)
		}
		config := NewConfig(uint32(len(out)))
		if v.Key != "" {
			key, err := hex.DecodeString(v.Key)
			if err != nil {
				t.Error(err)
			}
			config.Key = key
		}
		x, err := NewWithConfig(config)
		if err != nil {
			t.Fatalf("Error while creating blake2xb instance: %v", err)
		}
		x.Write(in)
		xof := make([]byte, len(out))
		_, err = io.ReadFull(x, xof)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(xof, out) {
			t.Fatalf("Failure for input %x: expected:\n%x\ngot\n%x", in, out, xof)
		}
	}
	dec.Token()
}
