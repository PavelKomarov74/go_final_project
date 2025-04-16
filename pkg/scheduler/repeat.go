package scheduler

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func AfterNow(date, now time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := now.Date()
	return y1 > y2 || (y1 == y2 && (m1 > m2 || (m1 == m2 && d1 > d2)))
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("пустое правило повторения")
	}

	startDate, err := time.Parse("20060102", dstart)
	if err != nil {
		return "", fmt.Errorf("неправильная дата начала: %w", err)
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", errors.New("неверный формат правила повторения")
	}

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("формат d должен быть: d <число>")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("некорректное значение в d (должно быть 1–400)")
		}
		date := startDate
		for {
			date = date.AddDate(0, 0, days)
			if AfterNow(date, now) {
				break
			}
		}
		return date.Format("20060102"), nil

	case "y":
		date := startDate
		for {
			date = date.AddDate(1, 0, 0)
			if AfterNow(date, now) {
				break
			}
		}
		return date.Format("20060102"), nil

	default:
		return "", fmt.Errorf("неподдерживаемое правило: %s", parts[0])
	}
}
