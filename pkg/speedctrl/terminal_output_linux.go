//go:build linux

package speedctrl

import "golang.org/x/sys/unix"

func enableOutputPostProcessing(fd int) error {
	state, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return err
	}
	state.Oflag |= unix.OPOST | unix.ONLCR
	return unix.IoctlSetTermios(fd, unix.TCSETS, state)
}
