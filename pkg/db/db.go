package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var (
	DB *sql.DB // Глобальная переменная с открытой БД
)

const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    repeat VARCHAR(128) NOT NULL DEFAULT ''
);
CREATE INDEX idx_date ON scheduler(date);
`

func Init(dbFile string) error {
	// Проверяем, существует ли файл
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	// Открываем подключение
	database, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("не удалось открыть БД: %w", err)
	}

	// Пробуем подключение
	if err = database.Ping(); err != nil {
		return fmt.Errorf("не удалось подключиться к БД: %w", err)
	}

	// Сохраняем ссылку в глобальной переменной
	DB = database

	// Если базы не было — создаём таблицу
	if install {
		_, err := DB.Exec(schema)
		if err != nil {
			return fmt.Errorf("ошибка при создании таблицы: %w", err)
		}
		fmt.Println("База данных и таблица успешно созданы.")
	} else {
		fmt.Println("База данных найдена и подключена.")
	}

	return nil
}

// Tasks возвращает список ближайших задач
func Tasks(limit int) ([]*Task, error) {
	rows, err := DB.Query(`
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
        ORDER BY date ASC 
        LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос задач: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var id int64
		var task Task
		if err := rows.Scan(&id, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("ошибка при чтении строки: %w", err)
		}
		task.ID = fmt.Sprintf("%d", id) // Преобразование ID в строку
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке строк: %w", err)
	}

	// Гарантируем, что возвращается пустой список, а не nil
	if tasks == nil {
		tasks = []*Task{}
	}

	return tasks, nil
}

func GetTask(id string) (*Task, error) {
	var task Task
	err := DB.QueryRow(`
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
        WHERE id = ?`, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to query task: %w", err)
	}
	return &task, nil
}

func UpdateTask(task *Task) error {
	query := `
        UPDATE scheduler 
        SET date = ?, title = ?, comment = ?, repeat = ? 
        WHERE id = ?`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check update result: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func UpdateDate(next string, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := DB.Exec(query, next, id)
	if err != nil {
		return fmt.Errorf("failed to update task date: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check update result: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check delete result: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
