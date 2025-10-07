package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
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
	}
}

func studnetsHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	port := ":3000"

	http.HandleFunc("/", rootHandler)

	http.HandleFunc("/teachers/", teacherHandler)

	http.HandleFunc("/students", studnetsHandler)

	fmt.Println("Server is running on port: ", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalln("Error starting server:", err)
	}
}
