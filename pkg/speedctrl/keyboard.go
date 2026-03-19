package speedctrl

import (
	"context"
	"io"
	"os"

	"golang.org/x/term"
)

var (
	keyboardIsTerminal                           = term.IsTerminal
	keyboardMakeRaw                              = term.MakeRaw
	keyboardRestore                              = term.Restore
	keyboardEnableOutputPostProcessing           = enableOutputPostProcessing
	keyboardFD                                   = func() int { return int(os.Stdin.Fd()) }
	keyboardInput                      io.Reader = os.Stdin
)

// StartKeyboardLoop enables raw-mode keyboard handling and toggles pause on space key.
func StartKeyboardLoop(ctx context.Context, c *Controller) error {
	fd := keyboardFD()
	isTerminal := keyboardIsTerminal
	makeRaw := keyboardMakeRaw
	restore := keyboardRestore
	input := keyboardInput

	if !isTerminal(fd) {
		return nil
	}

	oldState, err := makeRaw(fd)
	if err != nil {
		return err
	}
	if err := keyboardEnableOutputPostProcessing(fd); err != nil {
		_ = restore(fd, oldState)
		return err
	}

	go func() {
		defer func() {
			_ = restore(fd, oldState)
		}()
		buf := make([]byte, 1)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, readErr := input.Read(buf)
			if readErr != nil || n == 0 {
				return
			}
			if buf[0] == ' ' {
				c.ToggleManualPaused()
			}
		}
	}()

	return nil
}
