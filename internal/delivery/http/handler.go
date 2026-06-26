package http

import (
	"encoding/json"
	"net/http"

	"github.com/Ali506108/UserProfile-Service/internal/domain"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	useCase domain.UserUseCase
}

func RegistrationRouter(r *chi.Mux, uc domain.UserUseCase) {
	handler := &UserHandler{useCase: uc}

	r.Route("/api/users/v1/", func(r chi.Router) {
		r.Post("/", handler.Create)
		r.Get("/{id}", handler.Get)
	})
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var user domain.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if err := h.useCase.CreateUser(r.Context(), &user); err != nil {
		http.Error(w, "Error saving user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.useCase.GetUser(r.Context(), id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
