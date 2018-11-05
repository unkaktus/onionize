// +build !linux

package wslpath

import "errors"

var ErrNotImplemented = errors.New("wslpath: not implemented")

func FromWindows(p string) (string, error) {
	return "", ErrNotImplemented
}

func ToWindows(p string) (string, error) {
	return "", ErrNotImplemented
}
