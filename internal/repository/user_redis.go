package repository

import (
	"context"
	"encoding/json"

	"github.com/Ali506108/UserProfile-Service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type userRedisRepo struct {
	client *redis.Client
}

func NewUserRedisRepository(client *redis.Client) domain.UserRepository {
	return &userRedisRepo{client: client}
}

func (r *userRedisRepo) Save(ctx context.Context, user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, user.ID, data, 0).Err()

}

func (r *userRedisRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	data, err := r.client.Get(ctx, id).Result()

	if err != nil {
		return nil, err
	}

	var user domain.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}

	return &user, nil
}
