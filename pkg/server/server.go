package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"go1f/pkg/api"
)

func Run() error {
	port := 7540

	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	// Регистрируем API
	api.Init()

	// Обработчик статических файлов
	http.Handle("/", http.FileServer(http.Dir("web")))

	fmt.Printf("Сервер запущен на http://localhost:%d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
