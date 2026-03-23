package speedctrl

import "sync"

type Option func(*Controller)

func WithAPIEnabled(enabled bool) Option {
	return func(c *Controller) {
		if !enabled {
			c.apiPaused = false
		}
	}
}

type Controller struct {
	mu              sync.RWMutex
	apiPaused       bool
	manualPaused    bool
	speedMultiplier float64 // 0.0 to 1.0, applied as dispatch throttle
	gate            chan struct{}
}

// Speed multiplier bounds
const (
	speedMultiplierMin = 0.20 // 20% floor during ramp-down/ramp-up
	speedMultiplierMax = 1.00 // 100% ceiling
)

func NewController(opts ...Option) *Controller {
	c := &Controller{gate: make(chan struct{}), speedMultiplier: 1.0}
	close(c.gate)
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Controller) recomputeLocked() {
	paused := c.apiPaused || c.manualPaused
	if paused {
		select {
		case <-c.gate:
			c.gate = make(chan struct{})
		default:
		}
		return
	}
	select {
	case <-c.gate:
	default:
		close(c.gate)
	}
}

func (c *Controller) SetAPIPaused(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.apiPaused = v
	c.recomputeLocked()
}

func (c *Controller) SetManualPaused(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.manualPaused = v
	c.recomputeLocked()
}

func (c *Controller) ToggleManualPaused() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.manualPaused = !c.manualPaused
	c.recomputeLocked()
	return c.manualPaused
}

func (c *Controller) ManualPaused() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.manualPaused
}

func (c *Controller) APIPaused() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.apiPaused
}

func (c *Controller) IsPaused() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.apiPaused || c.manualPaused
}

func (c *Controller) Gate() <-chan struct{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.gate
}

// SetSpeedMultiplier sets speed multiplier directly.
func (c *Controller) SetSpeedMultiplier(v float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.speedMultiplier = v
}

// GetSpeedMultiplier returns the current speed multiplier.
func (c *Controller) GetSpeedMultiplier() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.speedMultiplier
}

// AdjustSpeedMultiplier increments multiplier by delta, clamped to [min, max].
// delta should be positive for ramp-up, negative for ramp-down.
// Does not affect pause state.
func (c *Controller) AdjustSpeedMultiplier(delta float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.speedMultiplier += delta
	if c.speedMultiplier < speedMultiplierMin {
		c.speedMultiplier = speedMultiplierMin
	}
	if c.speedMultiplier > speedMultiplierMax {
		c.speedMultiplier = speedMultiplierMax
	}
}

// ResetSpeedMultiplier resets to 0.0 — called when entering pause zone.
func (c *Controller) ResetSpeedMultiplier() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.speedMultiplier = 0.0
}
