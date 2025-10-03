package main

import (
	"log"
	"net/http"
	"tamis-server/internal/api"
)

func main() {
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)

	log.Println("Server running on :5000")
	err := http.ListenAndServe(":5000", mux)
	if err != nil {
		log.Fatal(err)
	}
}
