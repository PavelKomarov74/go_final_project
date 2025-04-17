package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go1f/pkg/db"
	"go1f/pkg/scheduler"
)

// addTaskHandle обрабатывает запросы на добавление задачи
func addTaskHandle(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	// Десериализация JSON
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON"})
		return
	}

	// Проверка заголовка
	if task.Title == "" {
		writeJson(w, http.StatusBadRequest, map[string]any{"error": "title is required"})
		return
	}

	// Проверка даты
	if err := checkDate(&task); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	// Добавление задачи
	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]any{"error": "database error"})
		return
	}

	// Возвращаем ID добавленной задачи
	writeJson(w, http.StatusOK, map[string]any{"id": id})
}

// checkDate проверяет и корректирует дату в задаче
func checkDate(task *db.Task) error {
	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	t, err := time.Parse("20060102", task.Date)
	if err != nil {
		return fmt.Errorf("invalid date format: %s", task.Date)
	}

	if scheduler.AfterNow(now, t) {
		if task.Repeat == "" {
			task.Date = now.Format("20060102")
		} else {
			next, err := scheduler.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				return fmt.Errorf("invalid repeat rule: %s", task.Repeat)
			}
			task.Date = next
		}
	}

	return nil
}

// writeJson отправляет JSON-ответ
func writeJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
