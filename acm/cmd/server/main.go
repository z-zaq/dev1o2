package main

import (
	"acm/internal/database"
	"acm/internal/handlers"
	"acm/internal/repository"

	// "acm/internal/auth"
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
	transactionRepo := &repository.TransactionRepository{
		DB: db,
	}
	handlers.TransactionRepo = transactionRepo
	handlers.UserRepo = userRepo
	err = userRepo.CreateTable()
	if err != nil {
		log.Fatal(err)
	}
	err = transactionRepo.CreateTable()
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
	mux.HandleFunc("/logout", handlers.LogoutHandler)
	mux.HandleFunc("/deposit", handlers.DepositHandler)
	mux.HandleFunc("/withdraw", handlers.WithdrawHandler)
	mux.HandleFunc("/history", handlers.HistoryHandler)
	mux.HandleFunc("/admin", handlers.AdminHandler)

	log.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
