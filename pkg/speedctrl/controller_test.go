package speedctrl

import "testing"

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
