package db

type Task struct {
	ID      string `json:"id"`                // идентификатор (будет заполнен после вставки)
	Date    string `json:"date"`              // дата выполнения
	Title   string `json:"title"`             // заголовок задачи (обязательное поле)
	Comment string `json:"comment,omitempty"` // комментарий
	Repeat  string `json:"repeat,omitempty"`  // правило повторения
}

// AddTask добавляет задачу в таблицу scheduler и возвращает ID новой записи
func AddTask(task *Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}
