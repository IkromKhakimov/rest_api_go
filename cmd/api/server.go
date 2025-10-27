package main

import (
	"crypto/tls"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	mw "restapi/internal/api/middlewares"
	"restapi/internal/api/router"
	"restapi/pkg/utils"

	"github.com/joho/godotenv"
)

//go:embed .env
var envFile embed.FS

func loadEnvFromEmbeddedFile() {
	// Read the embedded .env file
	content, err := envFile.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error reading .env file: %v", err)
	}

	// Create a temp file to load the env vars
	tempFile, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error creating .env file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write content of the embedded .env file to the time file
	_, err = tempFile.Write(content)
	if err != nil {
		log.Fatalf("Error writing .env file: %v", err)
	}
	err = tempFile.Close()
	if err != nil {
		log.Fatalf("Error closing .env file: %v", err)
	}

	// Load env vars from the temp file
	err = godotenv.Load(tempFile.Name())
	if err != nil {
		log.Fatalf("Error closing .env file: %v", err)
	}
}

func main() {
	// Only in production, for running source
	//err := godotenv.Load()
	//if err != nil {
	//	return
	//}

	// Load environment variables from the embedded .env file
	loadEnvFromEmbeddedFile()

	fmt.Println("Environment variable CERT_FILE:", os.Getenv("CERT"))

	port := os.Getenv("API_PORT")

	//cert := "cert.pem"
	//key := "key.pem"

	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS10,
	}

	//rl := mw.NewRateLimiter(5, time.Minute)
	//
	hppOptions := mw.HPPOptions{
		CheckQuery:                  true,
		CheckBody:                   true,
		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
		Whitelist:                   []string{"sortBy", "sortOrder", "name", "age", "class"},
	}

	//secureMux := mw.Hpp(hppOptions)(rl.Middleware(mw.ResponseTimeMiddleware(mw.SecurityHeaders(mux))))
	//secureMux := utils.ApplyMiddlewares((mux, mw.Hpp(hppOptions), mw.Compression, mw.SecurityHeaders, mw.ResponseTimeMiddleware)
	router := router.MainRouter()
	jwtMiddleware := mw.MiddlewaresExcludePaths(mw.JWTMiddleware, "/execs/login", "/execs/forgotpassword", "/execs/resetpassword/reset/")
	secureMux := utils.ApplyMiddlewares(router, jwtMiddleware, mw.Hpp(hppOptions), mw.Compression, mw.SecurityHeaders, mw.ResponseTimeMiddleware)
	//secureMux := mw.SecurityHeaders(router)

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
