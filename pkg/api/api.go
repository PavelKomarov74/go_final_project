package api

import (
	"encoding/json"
	"go1f/pkg/db"
	"go1f/pkg/scheduler"
	"log"
	"net/http"
	"time"
)

func logServerError(message string, err error) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
}

// tasksHandler обрабатывает запросы для получения списка задач
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJson(w, http.StatusMethodNotAllowed, map[string]any{"error": "unsupported method"})
		return
	}

	// Получение списка задач из базы данных
	tasks, err := db.Tasks(50) // Максимум 50 записей
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]any{"error": "failed to fetch tasks"})
		return
	}

	// Убедиться, что возвращается пустой список, если задач нет
	if tasks == nil {
		tasks = []*db.Task{}
	}

	// Возвращаем список задач в формате JSON
	writeJson(w, http.StatusOK, map[string]any{"tasks": tasks})
}

// taskHandler обрабатывает запросы для задач
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandle(w, r)
	case http.MethodGet:
		getTaskHandle(w, r)
	case http.MethodPut:
		updateTaskHandle(w, r)
	case http.MethodDelete:
		deleteTaskHandle(w, r)
	default:
		writeJson(w, http.StatusMethodNotAllowed, map[string]any{"error": "unsupported method"})
	}
}

func getTaskHandle(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]any{"error": "Не указан идентификатор"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, http.StatusNotFound, map[string]any{"error": "Задача не найдена"})
		return
	}

	writeJson(w, http.StatusOK, task)
}

func updateTaskHandle(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON"})
		return
	}

	if task.ID == "" {
		writeJson(w, http.StatusBadRequest, map[string]any{"error": "Не указан идентификатор"})
		return
	}

	if task.Title == "" {
		writeJson(w, http.StatusBadRequest, map[string]any{"error": "title is required"})
		return
	}

	if err := checkDate(&task); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	if err := db.UpdateTask(&task); err != nil {
		logServerError("Failed to update task in database", err)
		writeJson(w, http.StatusInternalServerError, map[string]any{"error": "Internal server error"})
		return
	}

	writeJson(w, http.StatusOK, map[string]any{})
}

func taskDoneHandle(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]any{"error": "Не указан идентификатор"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, http.StatusNotFound, map[string]any{"error": "Задача не найдена"})
		return
	}

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			logServerError("Failed to delete task", err)
			writeJson(w, http.StatusInternalServerError, map[string]any{"error": "Internal server error"})
			return
		}

		writeJson(w, http.StatusOK, map[string]any{})
		return
	}

	nextDate, err := scheduler.NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]any{"error": "Ошибка при вычислении следующей даты"})
		return
	}

	if err := db.UpdateDate(nextDate, id); err != nil {
		logServerError("Failed to update task date", err)
		writeJson(w, http.StatusInternalServerError, map[string]any{"error": "Internal server error"})
		return
	}

	writeJson(w, http.StatusOK, map[string]any{})
}

func deleteTaskHandle(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]any{"error": "Не указан идентификатор"})
		return
	}

	if err := db.DeleteTask(id); err != nil {
		logServerError("Failed to delete task", err)
		writeJson(w, http.StatusInternalServerError, map[string]any{"error": "Internal server error"})
		return
	}

	writeJson(w, http.StatusOK, map[string]any{})
}

// Init регистрирует маршруты API
func Init() {
	http.HandleFunc("/api/nextdate", nextDayHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/api/task/done", taskDoneHandle)
}
