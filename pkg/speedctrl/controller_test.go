package speedctrl

import "testing"

func TestController_ORGatePauseResume(t *testing.T) {
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
