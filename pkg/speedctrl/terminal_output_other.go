//go:build !darwin && !linux

package speedctrl

func enableOutputPostProcessing(fd int) error {
	_ = fd
	return nil
}
