package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	mw "restapi/internal/api/middlewares"
	"strings"
	"time"
)

type user struct {
	Name string `json:"name"`
	Age  string `json:"age"`
	City string `json:"city"`
}

func rootHandler(w http.ResponseWriter, r *http.Request) {

}

func teacherHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Println(r.URL.Path)
		path := strings.TrimPrefix(r.URL.Path, "/teachers/")
		userID := strings.TrimSuffix(path, "/")

		fmt.Println(userID)
		fmt.Println(r.URL.Query())
		queryParams := r.URL.Query()
		sortby := queryParams.Get("sortby")
		key := queryParams.Get("key")
		sortorder := queryParams.Get("sortorder")

		fmt.Printf("Sort by: %v, Sort order: %v, Key: %v\n", sortby, sortorder, key)
	}
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {

}

func execsHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	port := ":3000"

	cert := "cert.pem"
	key := "key.pem"

	mux := http.NewServeMux()

	http.HandleFunc("/", rootHandler)

	http.HandleFunc("/teachers/", teacherHandler)

	http.HandleFunc("/students", studentsHandler)

	http.HandleFunc("/execs/", execsHandler)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	rl := mw.NewRateLimiter(5, time.Minute)

	server := &http.Server{
		Addr:      port,
		Handler:   rl.Middleware(mw.ResponseTimeMiddleware(mux)),
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port: ", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting server:", err)
	}
}
