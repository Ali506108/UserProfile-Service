package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Ali506108/UserProfile-Service/internal/config"
	delivery "github.com/Ali506108/UserProfile-Service/internal/delivery/http"
	"github.com/Ali506108/UserProfile-Service/internal/repository"
	"github.com/Ali506108/UserProfile-Service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	cfg := config.LoadConfig()

	redisClient, err := repository.NewRedisClient(context.Background(), cfg.RedisAddr)

	if err != nil {
		log.Fatalf("Could not connect to redis: %v", err)
	}

	userRepo := repository.NewUserRedisRepository(redisClient)
	userUseCase := service.NewUserUseCase(userRepo)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	delivery.RegistrationRouter(r, userUseCase)

	log.Println("Server running on the port 9434")
	if err := http.ListenAndServe(":9434", r); r != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
