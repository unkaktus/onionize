// textqr.go - write QR code as text.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to textqr, using the creative
// commons "CC0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package textqr

import (
	"bytes"
	"io"

	"rsc.io/qr"
)

type Level int

const (
	L Level = iota // 20% redundant
	M              // 38% redundant
	Q              // 55% redundant
	H              // 65% redundant
)

var lightBlocks map[string]string = map[string]string{
	"large-full":  "\u2588\u2588",
	"large-empty": "  ",

	"compact-upper": "\u2580",
	"compact-lower": "\u2584",
	"compact-full":  "\u2588",
	"compact-empty": " ",
}

var darkBlocks map[string]string = map[string]string{
	"large-empty": "\u2588\u2588",
	"large-full":  "  ",

	"compact-lower": "\u2580",
	"compact-upper": "\u2584",
	"compact-empty": "\u2588",
	"compact-full":  " ",
}

func writeLine(w io.Writer, line []byte) (int, error) {
	return w.Write(append(line, byte('\n')))
}

// Write encodes text into a QR code and writes it into w.
// The redundancy level is set by level argument.
// If compact is true, compact QR codes are produced, otherwise
// large ones (x4 of compact ones) are produced.
// If inverted is true, the code "pixels" will appear as unfilled
// on filled background, and inverse otherwise.
func Write(w io.Writer, text string, level Level, compact, inverted bool) (int, error) {
	written := 0
	blocks := lightBlocks
	if inverted {
		blocks = darkBlocks
	}
	code, err := qr.Encode(text, qr.Level(level))
	if err != nil {
		return written, err
	}

	line := make([]byte, 0, code.Size+2)
	if compact {
		for i := 0; i < code.Size+1; i += 2 {
			line = append(line, blocks["compact-full"]...)
			for j := 0; j < code.Size; j++ {
				up, low := 0, 0
				if !code.Black(i-1, j) {
					up = 1
				}
				if !code.Black(i, j) {
					low = 2
				}
				c := ""
				switch up + low {
				case 0:
					c = blocks["compact-empty"]
				case 1:
					c = blocks["compact-upper"]
				case 2:
					c = blocks["compact-lower"]
				case 3:
					c = blocks["compact-full"]
				}
				line = append(line, c...)
			}
			line = append(line, blocks["compact-full"]...)
			n, err := writeLine(w, line)
			written += n
			if err != nil {
				return written, err
			}
			line = line[:0]
		}
		// Write the last line, as the number of lines in QR code is always odd.
		// Note that lines are not inversions of each other as there should be
		// a border right under the QR in one case and no border in another.
		c := blocks["compact-upper"]
		if inverted {
			c = blocks["compact-full"]
		}
		n, err := writeLine(w, bytes.Repeat([]byte(c), code.Size+2))
		written += n
		if err != nil {
			return written, err
		}
	} else {
		for i := 0; i < code.Size+2; i++ {
			line = append(line, blocks["large-full"]...)
			for j := 0; j < code.Size; j++ {
				if code.Black(i-1, j) {
					line = append(line, blocks["large-empty"]...)
				} else {
					line = append(line, blocks["large-full"]...)
				}
			}
			line = append(line, blocks["large-full"]...)
			n, err := writeLine(w, line)
			written += n
			if err != nil {
				return written, err
			}
			line = line[:0]
		}
	}
	return written, nil
}
