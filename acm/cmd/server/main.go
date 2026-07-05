package main

import (
	"acm/internal/database"
	"acm/internal/handlers"
	"acm/internal/repository"
	"log"
	"net/http"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	userRepo := &repository.UserRepository{
		DB: db,
		// handlers.UserRepo = userRepo
	}
	handlers.UserRepo = userRepo
	err = userRepo.CreateTable()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database connected successfully")

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static", fs))
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/about", handlers.AboutHandler)
	mux.HandleFunc("/contact", handlers.ContactHandler)
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/register", handlers.RegisterHandler)
	mux.HandleFunc("/dashboard", handlers.DashboardHandler)

	log.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
