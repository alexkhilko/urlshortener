package repository

import (
	"github.com/redis/go-redis/v9"
	"context"
	"fmt"
)



type Repository interface {
	Set(ctx context.Context, key string, value string) error
	Find(ctx context.Context, key string) (string, error) 
	GetDel(ctx context.Context, key string) (string, error)
}

type ApplicationRepository struct {
	DB *redis.Client
}

func NewApplicationRepository(db *redis.Client) *ApplicationRepository {
	return &ApplicationRepository{DB: db}
}

func (repo *ApplicationRepository) Set(ctx context.Context, key string, value string) error {
	err := repo.DB.Set(ctx, key, value, 0).Err()
	if err != nil {
		fmt.Println("error setting key", err)
		return err
	}
	return nil

}

func (repo *ApplicationRepository) Find(ctx context.Context, key string) (string, error) {
	val, err := repo.DB.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return val, nil
}

func (repo *ApplicationRepository) GetDel(ctx context.Context, key string) (string, error) {
	val, err := repo.DB.GetDel(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return val, nil
}




