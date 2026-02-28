package main_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/redis/go-redis/v9"
)

func TestWithRedis(t *testing.T) {
	ctx := context.Background()
	redisC, err := testcontainers.Run(
		ctx, "redis:latest",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp"),
			wait.ForLog("Ready to accept connections"),
		),
	)
	testcontainers.CleanupContainer(t, redisC)
	require.NoError(t, err)

	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		t.Error(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	err = client.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get(ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Println("Error connecting to Redis:", err)
	} else {
		fmt.Println("Successfully connected to Redis!")
	}
}
