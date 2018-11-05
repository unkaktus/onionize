// +build linux

package wslpath

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/nogoegst/wslpath/mount"
)

var (
	ErrNotFound = errors.New("mount not found")
	ErrInvalid  = errors.New("invalid path")
)

const (
	windowsOSPathSeparator = "\\"
	unixOSPathSeparator    = "/"
)

func drvfsFilter(i *mount.Info) (skip, stop bool) {
	return i.Fstype != "drvfs", false
}

func splitVolumeTarget(p string) (string, string) {
	if len(p) < 2 {
		return "", p
	}
	if p[1] != ':' {
		return "", p
	}
	c := strings.ToUpper(string(p[0]))[0]
	if !('A' <= c && c <= 'Z') {
		return "", p
	}
	return string(c) + ":", p[2:]
}

func toUnixSlash(p string) string {
	return strings.Replace(p, windowsOSPathSeparator, unixOSPathSeparator, -1)
}

func toWindowsSlash(p string) string {
	return strings.Replace(p, unixOSPathSeparator, windowsOSPathSeparator, -1)
}

// Convert Windows path like "C:\Windows\System32" to
// a WSL one like "/mnt/c/Windows/System32".
func FromWindows(p string) (string, error) {
	volume, target := splitVolumeTarget(p)
	if volume == "" {
		return p, ErrInvalid
	}
	mounts, err := mount.GetMounts(drvfsFilter)
	if err != nil {
		return "", err
	}
	mountpoint := ""
	for _, mnt := range mounts {
		if mnt.Source == volume {
			mountpoint = mnt.Mountpoint
			break
		}
	}
	if mountpoint == "" {
		return "", ErrNotFound
	}
	wslPath := filepath.Join(mountpoint, toUnixSlash(target))
	return wslPath, nil
}

// Convert WSL path like "/mnt/c/Windows/System32" to
// a Windows one like "C:\Windows\System32".
func ToWindows(p string) (string, error) {
	mounts, err := mount.GetMounts(drvfsFilter)
	if err != nil {
		return "", err
	}
	source, target := "", ""
	for _, mnt := range mounts {
		if strings.HasPrefix(p, mnt.Mountpoint) {
			source = mnt.Source
			target = strings.TrimPrefix(p, mnt.Mountpoint)
			break
		}
	}
	if source == "" {
		return "", ErrNotFound
	}
	windowsPath := toWindowsSlash(filepath.Join(source, target))
	return windowsPath, nil
}
