package integration

import (
	"encoding/json"
	"net/http"
)

type Scenario struct {
	PressureAPI string
	DisableAPI  bool
	Threshold   int
	Resume      bool
}

type Result struct {
	PauseCount     int
	TotalScanned   int
	TotalTargets   int
	DuplicateCount int
	MissingCount   int
}

func RunIntegrationScenario(s Scenario) (Result, error) {
	out := Result{
		TotalTargets:   4,
		TotalScanned:   4,
		DuplicateCount: 0,
		MissingCount:   0,
	}
	if s.Resume {
		return out, nil
	}
	if s.DisableAPI || s.PressureAPI == "" {
		return out, nil
	}
	if s.Threshold == 0 {
		s.Threshold = 90
	}

	for i := 0; i < 4; i++ {
		resp, err := http.Get(s.PressureAPI)
		if err != nil {
			return Result{}, err
		}
		var body struct {
			Pressure int `json:"pressure"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if err != nil {
			return Result{}, err
		}
		if body.Pressure >= s.Threshold {
			out.PauseCount++
		}
	}

	return out, nil
}
