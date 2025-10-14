package router

import (
	"net/http"
	"restapi/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)

	mux.HandleFunc("GET /teachers/", handlers.TeacherHandler)
	mux.HandleFunc("POST /teachers/", handlers.TeacherHandler)
	mux.HandleFunc("PUT /teachers/", handlers.TeacherHandler)
	mux.HandleFunc("PATCH /teachers/", handlers.TeacherHandler)
	mux.HandleFunc("DELETE /teachers/", handlers.TeacherHandler)

	mux.HandleFunc("/students", handlers.StudentsHandler)

	mux.HandleFunc("/execs/", handlers.ExecsHandler)

	return mux
}
