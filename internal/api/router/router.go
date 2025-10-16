package router

import (
	"net/http"
)

func MainRouter() *http.ServeMux {
	tRouter := teachersRouter()
	sRouter := studentsRouter()
	tRouter.Handle("/", sRouter)

	return tRouter
}
