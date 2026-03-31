package speedctrl

import (
	"context"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/term"
)

var (
	keyboardIsTerminal                           = term.IsTerminal
	keyboardMakeRaw                              = term.MakeRaw
	keyboardRestore                              = term.Restore
	keyboardEnableOutputPostProcessing           = enableOutputPostProcessing
	keyboardFD                                   = func() int { return int(os.Stdin.Fd()) }
	keyboardInput                      io.Reader = os.Stdin
	keyboardSignalNotify                         = signal.Notify
	keyboardSignalStop                           = signal.Stop
	keyboardPauseSignal                os.Signal = syscall.Signal(15)
)

// StartKeyboardLoop enables raw-mode keyboard handling where Enter resumes manual pause.
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
	var restoreOnce sync.Once
	restoreTerminal := func() {
		restoreOnce.Do(func() {
			_ = restore(fd, oldState)
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
			switch buf[0] {
			case '\r', '\n':
				c.SetManualPaused(false)
			}
		}
	}()

	return nil
}

// StartPauseSignalLoop enables pause control via SIGTERM (signal 15).
func StartPauseSignalLoop(ctx context.Context, c *Controller) {
	sigCh := make(chan os.Signal, 1)
	pauseSignal := keyboardPauseSignal
	notify := keyboardSignalNotify
	stop := keyboardSignalStop
	notify(sigCh, pauseSignal)

	go func() {
		defer stop(sigCh)
		for {
			select {
			case <-ctx.Done():
				return
			case sig := <-sigCh:
				if sig == nil {
					continue
				}
				c.SetManualPaused(true)
			}
		}
	}()
}
