package speedctrl

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"syscall"
	"testing"
	"time"

	"golang.org/x/term"
)

func TestControllerGate_WhenNotPaused_IsOpenAndKeyboardLoopStarts(t *testing.T) {
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

func TestControllerToggleManualPaused_WhenCalled_TogglesState(t *testing.T) {
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

func TestStartKeyboardLoop_WhenSpacePressed_DoesNotToggleManualPause(t *testing.T) {
	oldIsTerminal := keyboardIsTerminal
	oldMakeRaw := keyboardMakeRaw
	oldRestore := keyboardRestore
	oldEnableOutputPostProcessing := keyboardEnableOutputPostProcessing
	oldFD := keyboardFD
	oldInput := keyboardInput
	t.Cleanup(func() {
		keyboardIsTerminal = oldIsTerminal
		keyboardMakeRaw = oldMakeRaw
		keyboardRestore = oldRestore
		keyboardEnableOutputPostProcessing = oldEnableOutputPostProcessing
		keyboardFD = oldFD
		keyboardInput = oldInput
	})

	keyboardIsTerminal = func(fd int) bool { return true }
	keyboardMakeRaw = func(fd int) (*term.State, error) { return &term.State{}, nil }
	keyboardRestore = func(fd int, state *term.State) error { return nil }
	keyboardEnableOutputPostProcessing = func(fd int) error { return nil }
	keyboardFD = func() int { return 0 }
	keyboardInput = bytes.NewBuffer([]byte{' '})

	c := NewController()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartKeyboardLoop(ctx, c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	if c.ManualPaused() {
		t.Fatal("expected space key to be ignored")
	}
}

func TestStartKeyboardLoop_WhenEnterPressed_ClearsManualPause(t *testing.T) {
	oldIsTerminal := keyboardIsTerminal
	oldMakeRaw := keyboardMakeRaw
	oldRestore := keyboardRestore
	oldEnableOutputPostProcessing := keyboardEnableOutputPostProcessing
	oldFD := keyboardFD
	oldInput := keyboardInput
	t.Cleanup(func() {
		keyboardIsTerminal = oldIsTerminal
		keyboardMakeRaw = oldMakeRaw
		keyboardRestore = oldRestore
		keyboardEnableOutputPostProcessing = oldEnableOutputPostProcessing
		keyboardFD = oldFD
		keyboardInput = oldInput
	})

	keyboardIsTerminal = func(fd int) bool { return true }
	keyboardMakeRaw = func(fd int) (*term.State, error) { return &term.State{}, nil }
	keyboardRestore = func(fd int, state *term.State) error { return nil }
	keyboardEnableOutputPostProcessing = func(fd int) error { return nil }
	keyboardFD = func() int { return 0 }
	keyboardInput = bytes.NewBuffer([]byte{'\r'})

	c := NewController()
	c.SetManualPaused(true)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartKeyboardLoop(ctx, c); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	if c.ManualPaused() {
		t.Fatal("expected enter key to resume from manual pause")
	}
}

func TestStartPauseSignalLoop_WhenSIGTERMReceived_PausesManually(t *testing.T) {
	oldNotify := keyboardSignalNotify
	oldStop := keyboardSignalStop
	oldPauseSignal := keyboardPauseSignal
	t.Cleanup(func() {
		keyboardSignalNotify = oldNotify
		keyboardSignalStop = oldStop
		keyboardPauseSignal = oldPauseSignal
	})

	var registered chan<- os.Signal
	var capturedSignals []os.Signal
	keyboardSignalNotify = func(ch chan<- os.Signal, sig ...os.Signal) {
		registered = ch
		capturedSignals = append([]os.Signal(nil), sig...)
	}

	stopCalled := false
	keyboardSignalStop = func(ch chan<- os.Signal) {
		if ch == registered {
			stopCalled = true
		}
	}
	keyboardPauseSignal = syscall.Signal(15)

	c := NewController()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	StartPauseSignalLoop(ctx, c)

	if len(capturedSignals) != 1 || capturedSignals[0] != syscall.Signal(15) {
		t.Fatalf("expected pause signal registration for signal 15, got %#v", capturedSignals)
	}

	registered <- syscall.Signal(15)
	time.Sleep(20 * time.Millisecond)
	if !c.ManualPaused() {
		t.Fatal("expected signal 15 to enable manual pause")
	}

	cancel()
	time.Sleep(20 * time.Millisecond)
	if !stopCalled {
		t.Fatal("expected signal channel to be unregistered after cancellation")
	}
}

func TestStartKeyboardLoop_WhenMakeRawFails_ReturnsError(t *testing.T) {
	oldIsTerminal := keyboardIsTerminal
	oldMakeRaw := keyboardMakeRaw
	oldEnableOutputPostProcessing := keyboardEnableOutputPostProcessing
	oldFD := keyboardFD
	t.Cleanup(func() {
		keyboardIsTerminal = oldIsTerminal
		keyboardMakeRaw = oldMakeRaw
		keyboardEnableOutputPostProcessing = oldEnableOutputPostProcessing
		keyboardFD = oldFD
	})

	keyboardIsTerminal = func(fd int) bool { return true }
	keyboardMakeRaw = func(fd int) (*term.State, error) { return nil, errors.New("raw failed") }
	keyboardEnableOutputPostProcessing = func(fd int) error { return nil }
	keyboardFD = func() int { return 0 }

	err := StartKeyboardLoop(context.Background(), NewController())
	if err == nil {
		t.Fatal("expected make raw error")
	}
}

func TestStartKeyboardLoop_WhenRawModeEnabled_EnablesOutputPostProcessing(t *testing.T) {
	oldIsTerminal := keyboardIsTerminal
	oldMakeRaw := keyboardMakeRaw
	oldRestore := keyboardRestore
	oldEnableOutputPostProcessing := keyboardEnableOutputPostProcessing
	oldFD := keyboardFD
	oldInput := keyboardInput
	t.Cleanup(func() {
		keyboardIsTerminal = oldIsTerminal
		keyboardMakeRaw = oldMakeRaw
		keyboardRestore = oldRestore
		keyboardEnableOutputPostProcessing = oldEnableOutputPostProcessing
		keyboardFD = oldFD
		keyboardInput = oldInput
	})

	keyboardIsTerminal = func(fd int) bool { return true }
	keyboardMakeRaw = func(fd int) (*term.State, error) { return &term.State{}, nil }
	keyboardRestore = func(fd int, state *term.State) error { return nil }
	keyboardFD = func() int { return 123 }
	keyboardInput = bytes.NewBuffer(nil)

	enabledFD := -1
	keyboardEnableOutputPostProcessing = func(fd int) error {
		enabledFD = fd
		return nil
	}

	if err := StartKeyboardLoop(context.Background(), NewController()); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if enabledFD != 123 {
		t.Fatalf("expected output post-processing to run on fd=123, got %d", enabledFD)
	}
}

func TestStartKeyboardLoop_WhenEnableOutputPostProcessingFails_RestoresStateAndReturnsError(t *testing.T) {
	oldIsTerminal := keyboardIsTerminal
	oldMakeRaw := keyboardMakeRaw
	oldRestore := keyboardRestore
	oldEnableOutputPostProcessing := keyboardEnableOutputPostProcessing
	oldFD := keyboardFD
	t.Cleanup(func() {
		keyboardIsTerminal = oldIsTerminal
		keyboardMakeRaw = oldMakeRaw
		keyboardRestore = oldRestore
		keyboardEnableOutputPostProcessing = oldEnableOutputPostProcessing
		keyboardFD = oldFD
	})

	keyboardIsTerminal = func(fd int) bool { return true }
	keyboardFD = func() int { return 7 }
	rawState := &term.State{}
	keyboardMakeRaw = func(fd int) (*term.State, error) { return rawState, nil }

	restoreCalled := false
	keyboardRestore = func(fd int, state *term.State) error {
		if fd == 7 && state == rawState {
			restoreCalled = true
		}
		return nil
	}
	keyboardEnableOutputPostProcessing = func(fd int) error { return errors.New("output flags failed") }

	err := StartKeyboardLoop(context.Background(), NewController())
	if err == nil {
		t.Fatal("expected output post-processing error")
	}
	if !restoreCalled {
		t.Fatal("expected restore to be called when output post-processing setup fails")
	}
}

type blockingReader struct {
	started chan struct{}
	release chan struct{}
}

func (r *blockingReader) Read(p []byte) (int, error) {
	select {
	case <-r.started:
	default:
		close(r.started)
	}
	<-r.release
	return 0, io.EOF
}

func TestStartKeyboardLoop_WhenContextCanceledDuringBlockingRead_RestoresTerminal(t *testing.T) {
	oldIsTerminal := keyboardIsTerminal
	oldMakeRaw := keyboardMakeRaw
	oldRestore := keyboardRestore
	oldEnableOutputPostProcessing := keyboardEnableOutputPostProcessing
	oldFD := keyboardFD
	oldInput := keyboardInput
	t.Cleanup(func() {
		keyboardIsTerminal = oldIsTerminal
		keyboardMakeRaw = oldMakeRaw
		keyboardRestore = oldRestore
		keyboardEnableOutputPostProcessing = oldEnableOutputPostProcessing
		keyboardFD = oldFD
		keyboardInput = oldInput
	})

	keyboardIsTerminal = func(fd int) bool { return true }
	keyboardMakeRaw = func(fd int) (*term.State, error) { return &term.State{}, nil }
	keyboardEnableOutputPostProcessing = func(fd int) error { return nil }
	keyboardFD = func() int { return 9 }

	reader := &blockingReader{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	keyboardInput = reader

	restored := make(chan struct{}, 1)
	keyboardRestore = func(fd int, state *term.State) error {
		select {
		case restored <- struct{}{}:
		default:
		}
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := StartKeyboardLoop(ctx, NewController()); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	select {
	case <-reader.started:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected keyboard loop to start blocking read")
	}

	cancel()

	select {
	case <-restored:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected restore to run after context cancellation")
	}

	close(reader.release)
}
