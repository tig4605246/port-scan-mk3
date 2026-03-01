package speedctrl

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/term"
)

func TestGateAndKeyboardLoop(t *testing.T) {
	c := NewController()
	select {
	case <-c.Gate():
	default:
		t.Fatal("expected open gate")
	}
	if err := StartKeyboardLoop(context.Background(), c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToggleManualPaused(t *testing.T) {
	c := NewController()
	if c.ManualPaused() {
		t.Fatal("expected manual paused false initially")
	}
	if paused := c.ToggleManualPaused(); !paused {
		t.Fatal("expected paused after toggle")
	}
	if !c.ManualPaused() {
		t.Fatal("expected manual paused true")
	}
}

func TestStartKeyboardLoop_RawModeSpaceToggle(t *testing.T) {
	oldIsTerminal := keyboardIsTerminal
	oldMakeRaw := keyboardMakeRaw
	oldRestore := keyboardRestore
	oldFD := keyboardFD
	oldInput := keyboardInput
	t.Cleanup(func() {
		keyboardIsTerminal = oldIsTerminal
		keyboardMakeRaw = oldMakeRaw
		keyboardRestore = oldRestore
		keyboardFD = oldFD
		keyboardInput = oldInput
	})

	keyboardIsTerminal = func(fd int) bool { return true }
	keyboardMakeRaw = func(fd int) (*term.State, error) { return &term.State{}, nil }
	keyboardRestore = func(fd int, state *term.State) error { return nil }
	keyboardFD = func() int { return 0 }
	keyboardInput = bytes.NewBuffer([]byte{' '})

	c := NewController()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartKeyboardLoop(ctx, c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	if !c.ManualPaused() {
		t.Fatal("expected space key to toggle manual pause")
	}
}

func TestStartKeyboardLoop_MakeRawError(t *testing.T) {
	oldIsTerminal := keyboardIsTerminal
	oldMakeRaw := keyboardMakeRaw
	oldFD := keyboardFD
	t.Cleanup(func() {
		keyboardIsTerminal = oldIsTerminal
		keyboardMakeRaw = oldMakeRaw
		keyboardFD = oldFD
	})

	keyboardIsTerminal = func(fd int) bool { return true }
	keyboardMakeRaw = func(fd int) (*term.State, error) { return nil, errors.New("raw failed") }
	keyboardFD = func() int { return 0 }

	err := StartKeyboardLoop(context.Background(), NewController())
	if err == nil {
		t.Fatal("expected make raw error")
	}
}
