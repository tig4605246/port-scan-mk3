package testkit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
)

type MockPressureAPI struct {
	srv *httptest.Server
}

func NewMockPressureAPI(values []int) *MockPressureAPI {
	if len(values) == 0 {
		values = []int{0}
	}
	idx := 0
	mu := sync.Mutex{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		v := values[idx]
		if idx < len(values)-1 {
			idx++
		}
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]int{"pressure": v})
	}))
	return &MockPressureAPI{srv: srv}
}

func (m *MockPressureAPI) URL() string {
	return m.srv.URL
}

func (m *MockPressureAPI) Close() {
	m.srv.Close()
}
