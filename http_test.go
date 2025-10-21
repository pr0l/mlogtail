package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleHealth)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("handler returned unexpected status: got %v want ok", response["status"])
	}

	if response["version"] != VERSION {
		t.Errorf("handler returned unexpected version: got %v want %v", response["version"], VERSION)
	}
}

func TestHandleStats(t *testing.T) {
	// Инициализация счетчиков
	cfg := &Config{cmd: "file"}
	PostfixParserInit(cfg)

	// Добавляем тестовые данные
	msgStatusCounters.lock()
	msgStatusCounters.counters["received"] = 100
	msgStatusCounters.counters["delivered"] = 95
	msgStatusCounters.counters["rejected"] = 5
	msgStatusCounters.unlock()

	req, err := http.NewRequest("GET", "/stats", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleStats)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response StatsResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response.Received != 100 {
		t.Errorf("Expected received=100, got %d", response.Received)
	}

	if response.Delivered != 95 {
		t.Errorf("Expected delivered=95, got %d", response.Delivered)
	}

	if response.Rejected != 5 {
		t.Errorf("Expected rejected=5, got %d", response.Rejected)
	}
}

func TestHandleCounter(t *testing.T) {
	// Инициализация счетчиков
	cfg := &Config{cmd: "file"}
	PostfixParserInit(cfg)

	msgStatusCounters.lock()
	msgStatusCounters.counters["received"] = 42
	msgStatusCounters.unlock()

	req, err := http.NewRequest("GET", "/counter/received", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCounter)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response CounterResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response.Counter != "received" {
		t.Errorf("Expected counter='received', got '%s'", response.Counter)
	}

	if response.Value != 42 {
		t.Errorf("Expected value=42, got %d", response.Value)
	}
}

func TestHandleCounterInvalid(t *testing.T) {
	req, err := http.NewRequest("GET", "/counter/invalid-counter", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCounter)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var response ErrorResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response.Error == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestHandleReset(t *testing.T) {
	// Инициализация счетчиков
	cfg := &Config{cmd: "file"}
	PostfixParserInit(cfg)

	msgStatusCounters.lock()
	msgStatusCounters.counters["received"] = 100
	msgStatusCounters.unlock()

	req, err := http.NewRequest("POST", "/reset", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleReset)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Проверяем, что счетчики сброшены
	msgStatusCounters.lock()
	if msgStatusCounters.counters["received"] != 0 {
		t.Errorf("Expected counter to be reset to 0, got %d", msgStatusCounters.counters["received"])
	}
	msgStatusCounters.unlock()
}

func TestHandleResetWrongMethod(t *testing.T) {
	req, err := http.NewRequest("GET", "/reset", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleReset)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestHandleStatsReset(t *testing.T) {
	// Инициализация счетчиков
	cfg := &Config{cmd: "file"}
	PostfixParserInit(cfg)

	msgStatusCounters.lock()
	msgStatusCounters.counters["received"] = 50
	msgStatusCounters.counters["delivered"] = 45
	msgStatusCounters.unlock()

	req, err := http.NewRequest("POST", "/stats_reset", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleStatsReset)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response StatsResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	// Проверяем, что вернулись старые значения
	if response.Received != 50 {
		t.Errorf("Expected received=50, got %d", response.Received)
	}

	if response.Delivered != 45 {
		t.Errorf("Expected delivered=45, got %d", response.Delivered)
	}

	// Проверяем, что счетчики сброшены
	msgStatusCounters.lock()
	if msgStatusCounters.counters["received"] != 0 {
		t.Errorf("Expected counter to be reset to 0, got %d", msgStatusCounters.counters["received"])
	}
	msgStatusCounters.unlock()
}
