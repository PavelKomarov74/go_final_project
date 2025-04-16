package api

import (
	"encoding/json"
	"go1f/pkg/db"
	"go1f/pkg/scheduler"
	"net/http"
	"time"
)

// tasksHandler обрабатывает запросы для получения списка задач
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
		return
	}

	// Получение списка задач из базы данных
	tasks, err := db.Tasks(50) // Максимум 50 записей
	if err != nil {
		writeJson(w, map[string]any{"error": "failed to fetch tasks"})
		return
	}

	// Убедиться, что возвращается пустой список, если задач нет
	if tasks == nil {
		tasks = []*db.Task{}
	}

	// Возвращаем список задач в формате JSON
	writeJson(w, map[string]any{"tasks": tasks})
}

// taskHandler обрабатывает добавление задачи
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
		http.Error(w, "unsupported method", http.StatusMethodNotAllowed)
	}
}

func getTaskHandle(w http.ResponseWriter, r *http.Request) {
	// Получение параметра id из URL
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]any{"error": "Не указан идентификатор"})
		return
	}

	// Получение задачи из базы данных
	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]any{"error": "Задача не найдена"})
		return
	}

	// Возврат задачи в JSON формате
	writeJson(w, task)
}

func updateTaskHandle(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	// Десериализация JSON
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, map[string]any{"error": "invalid JSON"})
		return
	}

	// Проверка идентификатора
	if task.ID == "" {
		writeJson(w, map[string]any{"error": "Не указан идентификатор"})
		return
	}

	// Проверка данных
	if task.Title == "" {
		writeJson(w, map[string]any{"error": "title is required"})
		return
	}

	if err := checkDate(&task); err != nil {
		writeJson(w, map[string]any{"error": err.Error()})
		return
	}

	// Обновление задачи в базе данных
	if err := db.UpdateTask(&task); err != nil {
		writeJson(w, map[string]any{"error": err.Error()})
		return
	}

	// Возврат успешного ответа
	writeJson(w, map[string]any{})
}

func taskDoneHandle(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]any{"error": "Не указан идентификатор"})
		return
	}

	// Получение задачи по идентификатору
	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]any{"error": "Задача не найдена"})
		return
	}

	// Если повтор отсутствует, удаляем задачу
	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJson(w, map[string]any{"error": "Ошибка при удалении задачи"})
			return
		}
		writeJson(w, map[string]any{}) // Успешное удаление
		return
	}

	// Вычисление следующей даты
	nextDate, err := scheduler.NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeJson(w, map[string]any{"error": "Ошибка при вычислении следующей даты"})
		return
	}

	// Обновление даты задачи
	if err := db.UpdateDate(nextDate, id); err != nil {
		writeJson(w, map[string]any{"error": "Ошибка при обновлении даты"})
		return
	}

	writeJson(w, map[string]any{}) // Успешное завершение
}

func deleteTaskHandle(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, map[string]any{"error": "Не указан идентификатор"})
		return
	}

	// Удаление задачи
	if err := db.DeleteTask(id); err != nil {
		writeJson(w, map[string]any{"error": "Ошибка при удалении задачи"})
		return
	}

	writeJson(w, map[string]any{}) // Успешное завершение
}

// Init регистрирует маршруты API
func Init() {
	http.HandleFunc("/api/nextdate", nextDayHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/api/task/done", taskDoneHandle)
}
