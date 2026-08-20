package main

import (
	"acm/internal/database"
	"acm/internal/handlers"
	"acm/internal/middleware"
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
	}

	transactionRepo := &repository.TransactionRepository{
		DB: db,
	}

	planRepo := &repository.PlanRepository{
		DB: db,
	}

	handlers.TransactionRepo = transactionRepo
	handlers.PlanRepo = planRepo
	handlers.UserRepo = userRepo

	err = userRepo.CreateTable()
	if err != nil {
		log.Fatal(err)
	}

	err = transactionRepo.CreateTable()
	if err != nil {
		log.Fatal(err)
	}

	err = planRepo.CreateTable()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database connected successfully")

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static", fs))

	// Public routes — no auth required.
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/about", handlers.AboutHandler)
	mux.HandleFunc("/contact", handlers.ContactHandler)
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/register", handlers.RegisterHandler)
	mux.HandleFunc("/logout", handlers.LogoutHandler)

	// Authenticated routes — require a valid session.
	mux.HandleFunc("/dashboard", middleware.RequireAuth(userRepo, handlers.DashboardHandler))
	mux.HandleFunc("/plans", middleware.RequireAuth(userRepo, handlers.PlansHandler))
	mux.HandleFunc("/deposit", middleware.RequireAuth(userRepo, handlers.DepositHandler))
	mux.HandleFunc("/withdraw", middleware.RequireAuth(userRepo, handlers.WithdrawHandler))
	mux.HandleFunc("/history", middleware.RequireAuth(userRepo, handlers.HistoryHandler))
	mux.HandleFunc("/profile", middleware.RequireAuth(userRepo, handlers.ProfileHandler))
	mux.HandleFunc("/transfer", middleware.RequireAuth(userRepo, handlers.TransferHandler))
	mux.HandleFunc("/profile/edit", middleware.RequireAuth(userRepo, handlers.EditProfileHandler))
	mux.HandleFunc("/delete-account", middleware.RequireAuth(userRepo, handlers.DeleteAccountHandler))
	mux.HandleFunc("/change-password", middleware.RequireAuth(userRepo, handlers.ChangePasswordHandler))

	// Admin-only route.
	mux.HandleFunc("/admin", middleware.RequireRole(userRepo, "admin", handlers.AdminHandler))
	mux.HandleFunc("/admin/plans", middleware.RequireRole(userRepo, "admin", handlers.AdminCreatePlanHandler))

	log.Println("Server started on http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}
