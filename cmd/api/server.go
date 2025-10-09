package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	mw "restapi/internal/api/middlewares"
	"strconv"
	"strings"
	"sync"
)

//type user struct {
//	Name string `json:"name"`
//	Age  string `json:"age"`
//	City string `json:"city"`
//}

type Teacher struct {
	ID        int
	FirstName string
	LastName  string
	Class     string
	Subject   string
}

var teachers = make(map[int]Teacher)
var mutex = &sync.Mutex{}
var nextID = 1

func init() {
	teachers[nextID] = Teacher{
		ID:        nextID,
		FirstName: "John",
		LastName:  "Doe",
		Class:     "9A",
		Subject:   "Math",
	}
	nextID++
	teachers[nextID] = Teacher{
		ID:        nextID,
		FirstName: "Jane",
		LastName:  "Smith",
		Class:     "10A",
		Subject:   "Algebra",
	}
	nextID++
	teachers[nextID] = Teacher{
		ID:        nextID,
		FirstName: "Jane",
		LastName:  "Smith",
		Class:     "11A",
		Subject:   "Biology",
	}
}

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")
	fmt.Println(idStr)

	if idStr == "" {
		firstName := r.URL.Query().Get("first_name")
		lastName := r.URL.Query().Get("last_name")

		teacherList := make([]Teacher, 0, len(teachers))
		for _, teacher := range teachers {
			if (firstName == "" || teacher.FirstName == firstName) && (lastName == "" || teacher.LastName == lastName) {
				teacherList = append(teacherList, teacher)
			}
		}
		response := struct {
			Status string    `json:"status"`
			Count  int       `json:"count"`
			Data   []Teacher `json:"data"`
		}{
			Status: "success",
			Count:  len(teachers),
			Data:   teacherList,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	// Handle Path parameter
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	teacher, exists := teachers[id]
	if !exists {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(teacher)

}

func rootHandler(w http.ResponseWriter, r *http.Request) {

}

func teacherHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTeachersHandler(w, r)
	}
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {

}

func execsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		fmt.Println("Query:", r.URL.Query())
		fmt.Println("name:", r.URL.Query().Get("name"))

		err := r.ParseForm()
		if err != nil {
			return
		}
		fmt.Println("Form from POST methods:", err)
	}
}

func main() {
	port := ":3000"

	cert := "cert.pem"
	key := "key.pem"

	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)

	mux.HandleFunc("/teachers/", teacherHandler)

	mux.HandleFunc("/students", studentsHandler)

	mux.HandleFunc("/execs/", execsHandler)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	//rl := mw.NewRateLimiter(5, time.Minute)
	//
	//hppOptions := mw.HPPOptions{
	//	CheckQuery:                  true,
	//	CheckBody:                   true,
	//	CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
	//	Whitelist:                   []string{"sortBy", "sortOrder", "name", "age", "class"},
	//}

	//secureMux := mw.Hpp(hppOptions)(rl.Middleware(mw.ResponseTimeMiddleware(mw.SecurityHeaders(mux))))
	//secureMux := applyMiddlewares(mux, mw.Hpp(hppOptions), mw.Compression, mw.SecurityHeaders, mw.ResponseTimeMiddleware)
	secureMux := mw.SecurityHeaders(mux)

	server := &http.Server{
		Addr:      port,
		Handler:   secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port: ", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting server:", err)
	}
}

type Middleware func(handler http.Handler) http.Handler

func ApplyMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
