package domain

import "context"

type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Age         int    `json:"age"`
}

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
}

type UserUseCase interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
}
