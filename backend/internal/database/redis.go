package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(address string) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr: address,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
