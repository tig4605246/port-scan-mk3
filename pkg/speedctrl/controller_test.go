package speedctrl

import (
	"testing"
	"time"
)

func TestController_WhenPauseFlagsChange_UpdatesGateState(t *testing.T) {
	c := NewController(WithAPIEnabled(false))

	c.SetManualPaused(true)
	if !c.IsPaused() {
		t.Fatal("expected paused when manual paused")
	}

	c.SetAPIPaused(true)
	c.SetManualPaused(false)
	if !c.IsPaused() {
		t.Fatal("expected paused when api paused")
	}

	c.SetAPIPaused(false)
	if c.IsPaused() {
		t.Fatal("expected resumed when both flags false")
	}
}

func TestController_APIPausedAccessor_ReflectsLatestState(t *testing.T) {
	c := NewController()

	if c.APIPaused() {
		t.Fatal("expected api pause to start false")
	}

	c.SetAPIPaused(true)

	if !c.APIPaused() {
		t.Fatal("expected api pause to reflect latest true state")
	}
}

func TestController_SpeedMultiplier_AdjustAndClamp(t *testing.T) {
	c := NewController()

	if got := c.GetSpeedMultiplier(); got != 1.0 {
		t.Fatalf("expected initial multiplier 1.0, got %.2f", got)
	}

	c.AdjustSpeedMultiplier(0.10)
	if got := c.GetSpeedMultiplier(); got != 1.0 {
		t.Fatalf("expected clamped to 1.0, got %.2f", got)
	}

	c.SetSpeedMultiplier(0.25)
	c.AdjustSpeedMultiplier(-0.10)
	if got := c.GetSpeedMultiplier(); got != 0.20 {
		t.Fatalf("expected floor 0.20, got %.2f", got)
	}

	c.ResetSpeedMultiplier()
	if got := c.GetSpeedMultiplier(); got != 0.0 {
		t.Fatalf("expected 0.0 after reset, got %.2f", got)
	}
}

func TestController_SpeedMultiplier_DoesNotAffectGate(t *testing.T) {
	c := NewController(WithAPIEnabled(false))

	c.ResetSpeedMultiplier()
	select {
	case <-c.Gate():
	default:
		t.Fatal("gate should be open when not paused")
	}

	c.SetAPIPaused(true)
	c.SetSpeedMultiplier(1.0)
	select {
	case <-c.Gate():
		t.Fatal("gate should block when paused")
	case <-time.After(10 * time.Millisecond):
	}
}
