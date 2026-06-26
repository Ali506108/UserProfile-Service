package service

import (
	"context"
	"errors"

	"github.com/Ali506108/UserProfile-Service/internal/domain"
)

type userUseCase struct {
	useRepo domain.UserRepository
}

func NewUserUseCase(repo domain.UserRepository) domain.UserUseCase {
	return &userUseCase{useRepo: repo}
}

func (u *userUseCase) CreateUser(ctx context.Context, user *domain.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	return u.useRepo.Save(ctx, user)
}

func (u *userUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return u.useRepo.GetByID(ctx, id)
}
