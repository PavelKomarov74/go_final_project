package api

import (
	"fmt"
	"net/http"
	"time"

	"go1f/pkg/scheduler"
)

const dateLayout = "20060102"

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dstart := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	var err error

	if nowStr == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(dateLayout, nowStr)
		if err != nil {
			http.Error(w, "invalid now format", http.StatusBadRequest)
			return
		}
	}

	next, err := scheduler.NextDate(now, dstart, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, next)
}
