//go:build darwin

package speedctrl

import "golang.org/x/sys/unix"

func enableOutputPostProcessing(fd int) error {
	state, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return err
	}
	state.Oflag |= unix.OPOST | unix.ONLCR
	return unix.IoctlSetTermios(fd, unix.TIOCSETA, state)
}
