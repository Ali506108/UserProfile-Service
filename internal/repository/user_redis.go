package repository

import "github.com/redis/go-redis/v9"

type userRedisRepo struct {
	client *redis.Client
}
