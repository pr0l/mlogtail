package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

// StatsResponse структура для JSON-ответа
type StatsResponse struct {
	BytesReceived  uint64 `json:"bytes_received"`
	BytesDelivered uint64 `json:"bytes_delivered"`
	Received       uint64 `json:"received"`
	Delivered      uint64 `json:"delivered"`
	Forwarded      uint64 `json:"forwarded"`
	Deferred       uint64 `json:"deferred"`
	Bounced        uint64 `json:"bounced"`
	Rejected       uint64 `json:"rejected"`
	Held           uint64 `json:"held"`
	Discarded      uint64 `json:"discarded"`
	QueueSize      int    `json:"queue_size"`
}

// CounterResponse структура для JSON-ответа одного счетчика
type CounterResponse struct {
	Counter string `json:"counter"`
	Value   uint64 `json:"value"`
}

// ErrorResponse структура для JSON-ответа с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// getPostfixQueueSize возвращает количество писем в очереди Postfix
func getPostfixQueueSize() int {
	cmd := exec.Command("mailq")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	// Подсчитываем строки, начинающиеся с Queue ID (буквы и цифры)
	re := regexp.MustCompile(`(?m)^[0-9A-F]`)
	matches := re.FindAllString(string(output), -1)
	return len(matches)
}

// getStatsJSON возвращает все статистики в виде JSON
func getStatsJSON() StatsResponse {
	msgStatusCounters.lock()
	defer msgStatusCounters.unlock()

	return StatsResponse{
		BytesReceived:  msgStatusCounters.counters["bytes-received"],
		BytesDelivered: msgStatusCounters.counters["bytes-delivered"],
		Received:       msgStatusCounters.counters["received"],
		Delivered:      msgStatusCounters.counters["delivered"],
		Forwarded:      msgStatusCounters.counters["forwarded"],
		Deferred:       msgStatusCounters.counters["deferred"],
		Bounced:        msgStatusCounters.counters["bounced"],
		Rejected:       msgStatusCounters.counters["rejected"],
		Held:           msgStatusCounters.counters["held"],
		Discarded:      msgStatusCounters.counters["discarded"],
		QueueSize:      getPostfixQueueSize(),
	}
}

// getCounterJSON возвращает значение одного счетчика в виде JSON
func getCounterJSON(counter string) CounterResponse {
	msgStatusCounters.lock()
	defer msgStatusCounters.unlock()

	return CounterResponse{
		Counter: counter,
		Value:   msgStatusCounters.counters[counter],
	}
}

// resetCounters сбрасывает все счетчики
func resetCounters() {
	msgStatusCounters.lock()
	defer msgStatusCounters.unlock()
	msgStatusCounters.reset()
}

// handleStats обрабатывает запрос /stats
func handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := getStatsJSON()
	json.NewEncoder(w).Encode(stats)
}

// handleCounter обрабатывает запрос /counter/{name}
func handleCounter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Получаем имя счетчика из URL
	path := strings.TrimPrefix(r.URL.Path, "/counter/")
	counter := strings.TrimSpace(path)

	// Проверяем, существует ли такой счетчик
	validCounter := false
	for _, name := range PostfixStatusNames {
		if counter == name {
			validCounter = true
			break
		}
	}

	if !validCounter {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: fmt.Sprintf("Unknown counter: %s", counter),
		})
		return
	}

	result := getCounterJSON(counter)
	json.NewEncoder(w).Encode(result)
}

// handleReset обрабатывает запрос /reset
func handleReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Method not allowed. Use POST",
		})
		return
	}

	resetCounters()
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Counters reset successfully",
	})
}

// handleStatsReset обрабатывает запрос /stats_reset
func handleStatsReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Method not allowed. Use POST",
		})
		return
	}

	stats := getStatsJSON()
	resetCounters()
	json.NewEncoder(w).Encode(stats)
}

// handleHealth обрабатывает запрос /health
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": VERSION,
	})
}

// startHTTPServer запускает HTTP сервер
func startHTTPServer(addr string) {
	http.HandleFunc("/stats", handleStats)
	http.HandleFunc("/counter/", handleCounter)
	http.HandleFunc("/reset", handleReset)
	http.HandleFunc("/stats_reset", handleStatsReset)
	http.HandleFunc("/health", handleHealth)

	fmt.Printf("Starting HTTP server on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("HTTP server error: %s\n", err)
	}
}
