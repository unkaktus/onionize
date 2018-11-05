// tor.go - start an instance of little-t-tor.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to this module of onionize, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

package onionize

import (
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

const torrc = `RunAsDaemon 0
SOCKSPort 0
ControlPort 9999
AvoidDiskWrites 1
`

func runTor(ready chan<- struct{}) error {
	cmd := exec.Command("tor", "-f", "-")
	cmd.Stdin = strings.NewReader(torrc)
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	t := time.NewTicker(500 * time.Millisecond)
	for range t.C {
		c, err := net.Dial("tcp", "127.0.0.1:9999")
		if err == nil {
			c.Close()
			t.Stop()
			ready <- struct{}{}
			break
		}
	}
	return cmd.Wait()
}
