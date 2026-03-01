package testcontainers_test

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/redis/go-redis/v9"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
)

func setupRedisAndClient(t testing.TB, ctx context.Context) *redis.Client {
	redisC, err := testcontainers.Run(
		ctx, "redis:latest",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp"),
			wait.ForLog("Ready to accept connections"),
		),
	)
	testcontainers.CleanupContainer(t, redisC)
	if err != nil {
		t.Fatalf("Error=%v", err)
	}

	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("Error=%v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})
	return client
}

func TestRedisBasic(t *testing.T) {
	ctx := context.Background()
	client := setupRedisAndClient(t, ctx)

	err := client.Ping(ctx).Err()
	if err != nil {
		t.Fatalf("Error connecting to Redis: %v", err)
	} else {
		log.Println("Successfully connected to Redis!")
	}

	err = client.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		t.Fatalf("Error=%v", err)
	}

	val, err := client.Get(ctx, "key").Result()
	if err != nil {
		t.Fatalf("Error=%v", err)
	}
	if val != "value" {
		t.Fatalf("Wrong value '%s'", val)
	}

	err = client.Do(ctx, "set", "key", "changed_nx", "keepttl", "nx").Err()
	if err != redis.Nil {
		t.Fatalf("Error=%v", err)
	}

	val, err = client.Get(ctx, "key").Result()
	if err != nil {
		t.Fatalf("Error=%v", err)
	}

	if val != "value" {
		t.Fatalf("Wrong value '%s'", val)
	}
}

func TestRedisRedlock(t *testing.T) {
	ctx := context.Background()
	client := setupRedisAndClient(t, ctx)

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	wg := sync.WaitGroup{}

	startTime := time.Now().UTC()
	mutex := rs.NewMutex("redsync-mutex-key", redsync.WithExpiry(2*time.Second))
	err := mutex.LockContext(ctx)
	log.Printf("mutex.LockContext(ctx)=%v", err)
	if err != nil {
		t.Fatalf("mutex.LockContext(ctx) failed = %v", err)
	}

	wg.Add(1)
	go func() {
		sstartTime := time.Now().UTC()
		smutex := rs.NewMutex("redsync-mutex-key", redsync.WithExpiry(2*time.Second))
		err := smutex.LockContext(ctx)
		log.Printf("smutex.LockContext(ctx)=%v", err)
		if err != nil {
			panic(err)
		}
		state, err := smutex.UnlockContext(ctx)
		log.Printf("smutex.UnlockContext(ctx)=%v, %v [%v]", state, err, time.Since(sstartTime))
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	time.Sleep(time.Second)

	state, err := mutex.UnlockContext(ctx)
	log.Printf("mutex.UnlockContext(ctx)=%v, %v [%v]", state, err, time.Since(startTime))
	if err != nil {
		t.Fatalf("mutex.UnlockContext(ctx) failed = %v", err)
	}

	wg.Wait()
	log.Printf("Done [%v]", time.Since(startTime))
}

func TestRedisRedlockExpire(t *testing.T) {
	ctx := context.Background()
	client := setupRedisAndClient(t, ctx)

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	wg := sync.WaitGroup{}

	startTime := time.Now().UTC()
	mutex := rs.NewMutex("redsync-mutex-key", redsync.WithExpiry(time.Second))
	err := mutex.LockContext(ctx)
	log.Printf("mutex.LockContext(ctx)=%v", err)
	if err != nil {
		t.Fatalf("mutex.LockContext(ctx) failed = %v", err)
	}

	wg.Add(1)
	go func() {
		sstartTime := time.Now().UTC()
		smutex := rs.NewMutex("redsync-mutex-key", redsync.WithExpiry(time.Second))
		err := smutex.LockContext(ctx)
		log.Printf("smutex.LockContext(ctx)=%v", err)
		if err != nil {
			panic(err)
		}
		state, err := smutex.UnlockContext(ctx)
		log.Printf("smutex.UnlockContext(ctx)=%v, %v [%v]", state, err, time.Since(sstartTime))
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	time.Sleep(2 * time.Second)

	state, err := mutex.UnlockContext(ctx)
	log.Printf("mutex.UnlockContext(ctx)=%v, %v [%v]", state, err, time.Since(startTime))
	if err == nil {
		t.Fatalf("mutex.UnlockContext(ctx) success, wrong = %v", state)
	}

	wg.Wait()
	log.Printf("Done [%v]", time.Since(startTime))
}
