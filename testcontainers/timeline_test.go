package testcontainers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const parallelClients = 1900

type SourceRecord struct {
	StartTS int64  `json:"s"`
	EndTS   int64  `json:"e"`
	ID      string `json:"id"`
}

type SourceObject struct {
	ID      string         `json:"id"`
	Records []SourceRecord `json:"r"`
}

type ContainerStats struct {
	//Name     string `json:"name"`
	//ID       string `json:"id"`
	CPUStats struct {
		CPUUsage struct {
			Total  int64 `json:"total_usage"`
			Kernel int64 `json:"usage_in_kernelmode"`
			User   int64 `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemUsage int64 `json:"system_cpu_usage"`
		OnloneCPUs  int   `json:"online_cpus"`
	} `json:"cpu_stats"`
	MemoryStats struct {
		Usage int64 `json:"usage"`
		Limit int64 `json:"limit"`
		Stats struct {
			ActiveAnon  int64 `json:"active_anon"`
			Anon        int64 `json:"anon"`
			KernelStack int64 `json:"kernel_stack"`
			PGFaults    int64 `json:"pgfault"`
		} `json:"stats"`
	} `json:"memory_stats"`
}

func DumpContainerStats(tb testing.TB, ctx context.Context, totalCases int, actualClients int, dockerClient client.APIClient, containerID string) {
	stats, err := dockerClient.ContainerStats(ctx, containerID, false)
	if err != nil {
		tb.Fatalf("dockerClient.ContainerStats()=%v", err)
	}
	defer func() {
		err := stats.Body.Close()
		if err != nil {
			log.Printf("stats.Body.Close()=%v", err)
		}
	}()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, stats.Body)
	if err != nil {
		tb.Fatalf("io.Copy()=%v", err)
	}

	//log.Printf("stat=%s", buf.String())

	var containerStats ContainerStats
	err = json.Unmarshal(buf.Bytes(), &containerStats)
	if err != nil {
		tb.Fatalf("Error=%v", err)
	}
	log.Printf("Stats(%d/%d)=%v", totalCases, actualClients, containerStats)
}

func CalcIterations(cases int, totalClients int, client int) int {
	var result = cases / totalClients
	left := cases - result*totalClients
	if client < left {
		result += 1
	}
	return result
}

func TimelineSelectWithLua(ctx context.Context, keyPrefix string, start int64, end int64, client *redis.Client, script *redis.Script) (string, error) {
	targetKey := keyPrefix + "_target"
	sourceKey := keyPrefix + "_source"
	res, err := script.Run(ctx, client, []string{targetKey, sourceKey}, start, end).Result()
	//log.Printf("[%d-%d] res=%v, err=%v", start, end, res, err)
	if err != nil {
		return "", err
	}
	return res.(string), nil
}

func TimelineSelectWithRedlock(ctx context.Context, keyPrefix string, start int64, end int64, client *redis.Client, rs *redsync.Redsync) (string, error) {
	lockKey := keyPrefix + "_lock"
	mutex := rs.NewMutex(lockKey, redsync.WithExpiry(2*time.Second))
	err := mutex.LockContext(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		state, err := mutex.UnlockContext(ctx)
		if err != nil {
			log.Printf("mutex.UnlockContext(ctx)=%v,%v", state, err)
		}
	}()

	targetKey := keyPrefix + "_target"
	targetData, err := client.Get(ctx, targetKey).Result()
	if err == nil {
		return targetData, nil
	}

	sourceKey := keyPrefix + "_source"
	sourceData, err := client.Get(ctx, sourceKey).Result()
	if err != nil {
		return "", err
	}

	var source SourceObject
	err = json.Unmarshal([]byte(sourceData), &source)
	if err != nil {
		return "", err
	}

	targetData = ""

	for _, rec := range source.Records {
		if start < rec.EndTS && end >= rec.StartTS {
			targetDataBytes, err := json.Marshal(rec)
			if err != nil {
				return "", err
			}
			targetData = string(targetDataBytes)
			break
		}
	}

	err = client.Set(ctx, targetKey, targetData, 0).Err()
	if err != nil {
		return "", err
	}
	return targetData, nil
}

func TimelineSelectWithSetNX(ctx context.Context, keyPrefix string, start int64, end int64, client *redis.Client) (string, error) {
	targetKey := keyPrefix + "_target"
	targetData, err := client.Get(ctx, targetKey).Result()
	if err == nil {
		return targetData, nil
	}

	sourceKey := keyPrefix + "_source"
	sourceData, err := client.Get(ctx, sourceKey).Result()
	if err != nil {
		return "", err
	}

	var source SourceObject
	err = json.Unmarshal([]byte(sourceData), &source)
	if err != nil {
		return "", err
	}

	targetData = ""

	for _, rec := range source.Records {
		if start < rec.EndTS && end >= rec.StartTS {
			targetDataBytes, err := json.Marshal(rec)
			if err != nil {
				return "", err
			}
			targetData = string(targetDataBytes)
			break
		}
	}

	err = client.Do(ctx, "set", targetKey, targetData, "keepttl", "nx").Err()
	if err != nil && err != redis.Nil {
		return "", err
	}
	return targetData, nil
}

func GetScript() *redis.Script {
	return redis.NewScript(`
	local trgt_json = redis.call("GET", KEYS[1])
	if trgt_json then
		return trgt_json
	end

	local src_json = redis.call("GET", KEYS[2])
	if not src_json then
		return ""
	end
	trgt_json = ""
	local src = cjson.decode(src_json)
	local s = tonumber(ARGV[1])
	local e = tonumber(ARGV[2])
	for idx, rec in ipairs( src.r ) do
		if s < rec.e and e >= rec.s then
			trgt_json = cjson.encode(rec)
			break
		end
	end
	redis.call("SET", KEYS[1], trgt_json)
	return trgt_json
	`)
}

func CompareSourceRecords(recA string, recB string) bool {
	var srcA SourceRecord
	_ = json.Unmarshal([]byte(recA), &srcA)
	var srcB SourceRecord
	_ = json.Unmarshal([]byte(recB), &srcB)
	return srcA.ID == srcB.ID && srcA.StartTS == srcB.StartTS && srcA.EndTS == srcB.EndTS
}

func TestRedisTimelineSelectWithLuaInitial(t *testing.T) {
	ctx := context.Background()
	client := setupRedisAndClient(t, ctx)

	srcObj := SourceObject{ID: "obj#0", Records: []SourceRecord{
		{10000, 20000, "10000-20000"},
		{100000, 120000, "100000-120000"},
	}}

	srcData, err := json.Marshal(srcObj)
	if err != nil {
		t.Fatalf("Error=%v", err)
	}

	err = client.Set(ctx, "device0_source", srcData, 0).Err()
	if err != nil {
		t.Fatalf("Error=%v", err)
	}

	script := GetScript()

	res0, err := TimelineSelectWithLua(ctx, "device0", 0, 1000, client, script)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}
	if res0 != "" {
		t.Fatalf("Redis return wong value %s, expected empty", res0)
	}

	target0, err := client.Get(ctx, "device0_target").Result()
	//log.Printf("target0=%v, err=%v\n", target0, err)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}
	if target0 != "" {
		t.Fatalf("Wrong value '%s'. Expected ''", target0)
	}

	err = client.Del(ctx, "device0_target").Err()
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}

	res1, err := TimelineSelectWithLua(ctx, "device0", 11000, 12000, client, script)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}
	if !CompareSourceRecords(res1, "{\"e\":20000,\"s\":10000,\"id\":\"10000-20000\"}") {
		t.Fatalf("Redis return wong value %s, expected {\"e\":20000,\"s\":10000,\"id\":\"10000-20000\"}", res1)
	}

	target0, err = client.Get(ctx, "device0_target").Result()
	//log.Printf("target0=%v, err=%v\n", target0, err)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}

	if !CompareSourceRecords(target0, res1) {
		t.Fatalf("Redis return wong value %s, expected %s", target0, res1)
	}
}

func TestRedisTimelineSelectWithRedlockInitial(t *testing.T) {
	ctx := context.Background()
	client := setupRedisAndClient(t, ctx)

	srcObj := SourceObject{ID: "obj#0", Records: []SourceRecord{
		{10000, 20000, "10000-20000"},
		{100000, 120000, "100000-120000"},
	}}

	srcData, err := json.Marshal(srcObj)
	if err != nil {
		t.Fatalf("Error=%v", err)
	}

	err = client.Set(ctx, "device0_source", srcData, 0).Err()
	if err != nil {
		t.Fatalf("Error=%v", err)
	}

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	res0, err := TimelineSelectWithRedlock(ctx, "device0", 0, 1000, client, rs)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}
	if res0 != "" {
		t.Fatalf("Redis return wong value %s, expected empty", res0)
	}

	target0, err := client.Get(ctx, "device0_target").Result()
	//log.Printf("target0=%v, err=%v\n", target0, err)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}
	if target0 != "" {
		t.Fatalf("Wrong value '%s'. Expected ''", target0)
	}

	err = client.Del(ctx, "device0_target").Err()
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}

	res1, err := TimelineSelectWithRedlock(ctx, "device0", 11000, 12000, client, rs)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}
	if !CompareSourceRecords(res1, "{\"e\":20000,\"s\":10000,\"id\":\"10000-20000\"}") {
		t.Fatalf("Redis return wong value %s, expected {\"e\":20000,\"s\":10000,\"id\":\"10000-20000\"}", res1)
	}

	target0, err = client.Get(ctx, "device0_target").Result()
	//log.Printf("target0=%v, err=%v\n", target0, err)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}

	if !CompareSourceRecords(target0, res1) {
		t.Fatalf("Redis return wong value %s, expected %s", target0, res1)
	}
}

func TestRedisTimelineSelectWithSetNXInitial(t *testing.T) {
	ctx := context.Background()
	client := setupRedisAndClient(t, ctx)

	srcObj := SourceObject{ID: "obj#0", Records: []SourceRecord{
		{10000, 20000, "10000-20000"},
		{100000, 120000, "100000-120000"},
	}}

	srcData, err := json.Marshal(srcObj)
	if err != nil {
		t.Fatalf("Error=%v", err)
	}

	err = client.Set(ctx, "device0_source", srcData, 0).Err()
	if err != nil {
		t.Fatalf("Error=%v", err)
	}

	res0, err := TimelineSelectWithSetNX(ctx, "device0", 0, 1000, client)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}
	if res0 != "" {
		t.Fatalf("Redis return wong value %s, expected empty", res0)
	}

	target0, err := client.Get(ctx, "device0_target").Result()
	//log.Printf("target0=%v, err=%v\n", target0, err)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}
	if target0 != "" {
		t.Fatalf("Wrong value '%s'. Expected ''", target0)
	}

	err = client.Del(ctx, "device0_target").Err()
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}

	res1, err := TimelineSelectWithSetNX(ctx, "device0", 11000, 12000, client)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}
	if !CompareSourceRecords(res1, "{\"e\":20000,\"s\":10000,\"id\":\"10000-20000\"}") {
		t.Fatalf("Redis return wong value %s, expected {\"e\":20000,\"s\":10000,\"id\":\"10000-20000\"}", res1)
	}

	target0, err = client.Get(ctx, "device0_target").Result()
	//log.Printf("target0=%v, err=%v\n", target0, err)
	if err != nil {
		t.Fatalf("Redis return error=%v", err)
	}

	if !CompareSourceRecords(target0, res1) {
		t.Fatalf("Redis return wong value %s, expected %s", target0, res1)
	}
}

func BenchmarkRedisTimelineSelectWithLua(b *testing.B) {
	ctx := context.Background()

	redisC, err := testcontainers.Run(
		ctx, "redis:latest",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp"),
			wait.ForLog("Ready to accept connections"),
		),
	)
	testcontainers.CleanupContainer(b, redisC)
	if err != nil {
		b.Fatalf("Error=%v", err)
	}

	provider, _ := testcontainers.NewDockerProvider()
	dockerClient := provider.Client()
	containerID := redisC.GetContainerID()

	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		b.Fatalf("Error=%v", err)
	}

	srcObj := SourceObject{ID: "obj#0", Records: []SourceRecord{
		{10000, 20000, "10000-20000"},
		{100000, 120000, "100000-120000"},
	}}

	srcData, err := json.Marshal(srcObj)
	if err != nil {
		b.Fatalf("Error=%v", err)
	}

	script := GetScript()

	actualClients := min(parallelClients, b.N)
	wg := sync.WaitGroup{}

	b.ResetTimer()

	for i := range actualClients {
		wg.Add(1)
		go func() {
			client := redis.NewClient(&redis.Options{
				Addr: endpoint,
			})

			key := fmt.Sprintf("device%d", i)
			keyTarget := key + "_target"
			keySource := key + "_source"

			err = client.Set(ctx, keySource, srcData, 0).Err()
			if err != nil {
				log.Printf("Error=%v", err)
				panic(err)
			}

			for j := range CalcIterations(b.N, actualClients, i) {
				_ = j

				_, err := TimelineSelectWithLua(ctx, key, 9000, 11000, client, script)
				if err != nil {
					log.Printf("Redis return error=%v", err)
					panic(err)
				}

				err = client.Del(ctx, keyTarget).Err()
				if err != nil {
					log.Printf("Redis return error=%v", err)
					panic(err)
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()

	DumpContainerStats(b, ctx, b.N, actualClients, dockerClient, containerID)
}

func BenchmarkRedisTimelineSelectWithRedlock(b *testing.B) {
	ctx := context.Background()

	redisC, err := testcontainers.Run(
		ctx, "redis:latest",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp"),
			wait.ForLog("Ready to accept connections"),
		),
	)
	testcontainers.CleanupContainer(b, redisC)
	if err != nil {
		b.Fatalf("Error=%v", err)
	}

	provider, _ := testcontainers.NewDockerProvider()
	dockerClient := provider.Client()
	containerID := redisC.GetContainerID()

	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		b.Fatalf("Error=%v", err)
	}

	srcObj := SourceObject{ID: "obj#0", Records: []SourceRecord{
		{10000, 20000, "10000-20000"},
		{100000, 120000, "100000-120000"},
	}}

	srcData, err := json.Marshal(srcObj)
	if err != nil {
		b.Fatalf("Error=%v", err)
	}

	actualClients := min(parallelClients, b.N)
	wg := sync.WaitGroup{}

	b.ResetTimer()

	for i := range actualClients {
		wg.Add(1)
		go func() {
			client := redis.NewClient(&redis.Options{
				Addr: endpoint,
			})

			pool := goredis.NewPool(client)
			rs := redsync.New(pool)

			key := fmt.Sprintf("device%d", i)
			keyTarget := key + "_target"
			keySource := key + "_source"

			err = client.Set(ctx, keySource, srcData, 0).Err()
			if err != nil {
				log.Printf("Error=%v", err)
				panic(err)
			}

			for j := range CalcIterations(b.N, actualClients, i) {
				_ = j

				_, err := TimelineSelectWithRedlock(ctx, key, 9000, 11000, client, rs)
				if err != nil {
					log.Printf("Redis return error=%v", err)
					panic(err)
				}

				err = client.Del(ctx, keyTarget).Err()
				if err != nil {
					log.Printf("Redis return error=%v", err)
					panic(err)
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()

	DumpContainerStats(b, ctx, b.N, actualClients, dockerClient, containerID)
}

func BenchmarkRedisTimelineSelectWithSetNX(b *testing.B) {
	ctx := context.Background()

	redisC, err := testcontainers.Run(
		ctx, "redis:latest",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp"),
			wait.ForLog("Ready to accept connections"),
		),
	)
	testcontainers.CleanupContainer(b, redisC)
	if err != nil {
		b.Fatalf("Error=%v", err)
	}

	provider, _ := testcontainers.NewDockerProvider()
	dockerClient := provider.Client()
	containerID := redisC.GetContainerID()

	endpoint, err := redisC.Endpoint(ctx, "")
	if err != nil {
		b.Fatalf("Error=%v", err)
	}

	srcObj := SourceObject{ID: "obj#0", Records: []SourceRecord{
		{10000, 20000, "10000-20000"},
		{100000, 120000, "100000-120000"},
	}}

	srcData, err := json.Marshal(srcObj)
	if err != nil {
		b.Fatalf("Error=%v", err)
	}

	actualClients := min(parallelClients, b.N)
	wg := sync.WaitGroup{}

	b.ResetTimer()

	for i := range actualClients {
		wg.Add(1)
		go func() {
			client := redis.NewClient(&redis.Options{
				Addr: endpoint,
			})

			key := fmt.Sprintf("device%d", i)
			keyTarget := key + "_target"
			keySource := key + "_source"

			err = client.Set(ctx, keySource, srcData, 0).Err()
			if err != nil {
				log.Printf("Error=%v", err)
				panic(err)
			}

			for j := range CalcIterations(b.N, actualClients, i) {
				_ = j

				_, err := TimelineSelectWithSetNX(ctx, key, 9000, 11000, client)
				if err != nil {
					log.Printf("Redis return error=%v", err)
					panic(err)
				}

				err = client.Del(ctx, keyTarget).Err()
				if err != nil {
					log.Printf("Redis return error=%v", err)
					panic(err)
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()

	DumpContainerStats(b, ctx, b.N, actualClients, dockerClient, containerID)
}
