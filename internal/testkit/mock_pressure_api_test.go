package testkit

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestMockPressureAPI_ServesSequence(t *testing.T) {
	api := NewMockPressureAPI([]int{10, 20})
	defer api.Close()

	resp1, err := http.Get(api.URL())
	if err != nil {
		t.Fatal(err)
	}
	defer resp1.Body.Close()
	var b1 map[string]int
	if err := json.NewDecoder(resp1.Body).Decode(&b1); err != nil {
		t.Fatal(err)
	}
	if b1["pressure"] != 10 {
		t.Fatalf("unexpected pressure: %d", b1["pressure"])
	}

	resp2, err := http.Get(api.URL())
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	var b2 map[string]int
	if err := json.NewDecoder(resp2.Body).Decode(&b2); err != nil {
		t.Fatal(err)
	}
	if b2["pressure"] != 20 {
		t.Fatalf("unexpected pressure: %d", b2["pressure"])
	}
}
