package speedctrl

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

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
		if err := restore(fd, oldState); err != nil {
			fmt.Fprintf(os.Stderr, "speedctrl: failed to restore terminal state: %v\n", err)
		}
		return err
	}
	var restoreOnce sync.Once
	restoreTerminal := func() {
		restoreOnce.Do(func() {
			if err := restore(fd, oldState); err != nil {
				fmt.Fprintf(os.Stderr, "speedctrl: failed to restore terminal state: %v\n", err)
			}
		})
	}

	go func() {
		<-ctx.Done()
		restoreTerminal()
	}()

	go func() {
		defer func() {
			restoreTerminal()
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
			select {
			case <-ctx.Done():
				return
			default:
			}
			if buf[0] == ' ' {
				c.ToggleManualPaused()
			}
		}
	}()

	return nil
}
