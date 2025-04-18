package main

import (
	"log"
	"os"

	"go1f/pkg/db"
	"go1f/pkg/server"
)

func main() {
	// Читаем путь к файлу базы из переменной окружения TODO_DBFILE
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	// Инициализация базы данных
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("Ошибка при инициализации базы данных: %v", err)
	}

	// Запуск сервера
	if err := server.Run(); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
