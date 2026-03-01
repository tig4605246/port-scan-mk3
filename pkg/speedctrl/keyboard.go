package speedctrl

import (
	"context"
	"io"
	"os"

	"golang.org/x/term"
)

var (
	keyboardIsTerminal           = term.IsTerminal
	keyboardMakeRaw              = term.MakeRaw
	keyboardRestore              = term.Restore
	keyboardFD                   = func() int { return int(os.Stdin.Fd()) }
	keyboardInput      io.Reader = os.Stdin
)

// StartKeyboardLoop enables raw-mode keyboard handling and toggles pause on space key.
func StartKeyboardLoop(ctx context.Context, c *Controller) error {
	fd := keyboardFD()
	if !keyboardIsTerminal(fd) {
		return nil
	}

	oldState, err := keyboardMakeRaw(fd)
	if err != nil {
		return err
	}

	go func() {
		defer func() {
			_ = keyboardRestore(fd, oldState)
		}()
		buf := make([]byte, 1)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			n, readErr := keyboardInput.Read(buf)
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
