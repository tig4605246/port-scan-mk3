package speedcontrol

import "sync"

type Collector struct {
	mu       sync.Mutex
	scenario string
	events   []Event
}

func NewCollector(scenario string) *Collector {
	return &Collector{scenario: scenario}
}

func (c *Collector) Record(e Event) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e.Scenario == "" {
		e.Scenario = c.scenario
	}
	c.events = append(c.events, e)
}

func (c *Collector) Events() []Event {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]Event, len(c.events))
	copy(out, c.events)
	return out
}
