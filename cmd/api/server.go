package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	mw "restapi/internal/api/middlewares"
	"restapi/internal/api/router"
)

func main() {
	port := ":3000"

	cert := "cert.pem"
	key := "key.pem"

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
	//secureMux := utils.ApplyMiddlewares((mux, mw.Hpp(hppOptions), mw.Compression, mw.SecurityHeaders, mw.ResponseTimeMiddleware)
	router := router.Router()
	secureMux := mw.SecurityHeaders(router)

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
