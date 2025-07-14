package main

import (
	"log"
	"net/http"

	"github.com/tkowalski/socgo/internal/server"
)

func main() {
	srv := server.New()
	
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", srv); err != nil {
		log.Fatal(err)
	}
}