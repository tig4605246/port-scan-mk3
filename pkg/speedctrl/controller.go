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
	mu           sync.RWMutex
	apiPaused    bool
	manualPaused bool
	gate         chan struct{}
}

func NewController(opts ...Option) *Controller {
	c := &Controller{gate: make(chan struct{})}
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
