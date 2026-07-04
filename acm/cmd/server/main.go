package main

import (
	"acm/internal/handlers"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static", fs))
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/about", handlers.AboutHandler)
	mux.HandleFunc("/contact", handlers.ContactHandler)
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/register", handlers.RegisterHandler)

	log.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
