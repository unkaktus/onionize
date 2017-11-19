// byteqr.go - write QR code bytes
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to byteqr, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package byteqr

import (
	"io"

	"rsc.io/qr"
)

const (
	WHITE = "\033[47m██\033[0m"
	BLACK = "\033[40m  \033[0m"
)

// Write encodes text into a QR code and writes it into w.
func Write(w io.Writer, text string, level qr.Level, white, black []byte) error {
	code, err := qr.Encode(text, level)
	if err != nil {
		return err
	}
	if white == nil || black == nil {
		white, black = []byte(WHITE), []byte(BLACK)
	}

	line := make([]byte, code.Size+2)
	for i := 0; i < code.Size+2; i++ {
		line = append(line, white...)
		for j := 0; j < code.Size; j++ {
			if code.Black(i-1, j) {
				line = append(line, black...)
			} else {
				line = append(line, white...)
			}
		}
		line = append(line, white...)
		_, err := w.Write(append(line, byte('\n')))
		if err != nil {
			return err
		}
		line = line[:0]
	}
	return nil
}
