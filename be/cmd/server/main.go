package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Starting Bahasa Daerah Learning Platform server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
