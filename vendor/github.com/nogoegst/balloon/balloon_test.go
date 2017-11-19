// balloon_test.go - test the Balloon implementation.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of blake2xb, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package balloon

import (
	"crypto/rand"
	"crypto/sha512"
	"testing"
)

func BenchmarkBalloon(b *testing.B) {
	ps := make([]byte, 8+8)
	rand.Read(ps)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Balloon(sha512.New(), ps[:8], ps[8:], 16, 16)
	}
}

func BenchmarkBalloonM(b *testing.B) {
	ps := make([]byte, 8+8)
	rand.Read(ps)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BalloonM(sha512.New, ps[:8], ps[8:], 16, 16, 4)
	}
}
