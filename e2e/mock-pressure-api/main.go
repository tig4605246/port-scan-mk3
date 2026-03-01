package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	mode := getenv("MODE", "ok")
	addr := getenv("ADDR", ":8080")
	pressure := getenvInt("PRESSURE", 20)
	delayMS := getenvInt("DELAY_MS", 5000)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/pressure", newPressureHandler(mode, pressure, delayMS))

	log.Printf("mock pressure api listening on %s mode=%s", addr, mode)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func newPressureHandler(mode string, pressure, delayMS int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch mode {
		case "ok":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]int{"pressure": pressure})
		case "fail":
			http.Error(w, "mock fail", http.StatusInternalServerError)
		case "timeout":
			time.Sleep(time.Duration(delayMS) * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]int{"pressure": pressure})
		default:
			http.Error(w, "unknown mode", http.StatusInternalServerError)
		}
	}
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
