// cmd/api/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/Oladele-israel/socialmedia-post-automation/internal/auth"
	"github.com/Oladele-israel/socialmedia-post-automation/internal/middleware"
	"github.com/Oladele-israel/socialmedia-post-automation/pkg/cache"
	"github.com/Oladele-israel/socialmedia-post-automation/pkg/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Infrastructure
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := cache.NewRedisClient()
	if err != nil {
		log.Fatal(err)
	}

	// Auth middleware — built once, passed to any router that needs it
	authMiddleware := middleware.RequireAuth(redisClient)

	// Wire up auth
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, redisClient)
	authHandler := auth.NewHandler(authService)

	// Main router — only global middleware here
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Mount each service router — main.go never sees individual routes
	r.Mount("/api/auth", authHandler.Router(authMiddleware))
	// Coming soon:
	// r.Mount("/api/posts", postsHandler.Router(authMiddleware))
	// r.Mount("/api/scheduler", schedulerHandler.Router(authMiddleware))

	port := os.Getenv("APP_PORT")
	log.Printf("🚀 Server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMmJhMTY0NGMtYjU2YS00NzQ5LThlYTctNDE4YTBlOGViMDYyIiwiZW1haWwiOiJvbW9sZW9sYWRlbGU0M0BnbWFpbC5jb20iLCJleHAiOjE3Nzk5NTM1ODIsImlhdCI6MTc3OTg2NzE4MiwianRpIjoiMTdlZDVlOGYtOWMxYS00MzJhLWIwMGYtMjMyMjZiZmQzODU5In0.OBdCT36xTEXgAWRL5v-XBUHhYAjkaM0c08A68lkpxow",
// "access_token_id": "17ed5e8f-9c1a-432a-b00f-23226bfd3859",
// "refresh_token": "2a740fcb-863a-4e21-9bf8-db2eaf9cac1c"
