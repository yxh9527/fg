package dao

import (
	"context"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestInitializeRedisConfigsImportsSnapshotOnce(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	defaults := defaultRedisConfigs
	if len(defaults) != 147 {
		t.Fatalf("snapshot has %d keys, want 147", len(defaults))
	}
	inserted, err := initializeRedisConfigs(ctx, client)
	if err != nil || inserted != len(defaults) {
		t.Fatalf("first import: inserted=%d err=%v", inserted, err)
	}
	for key, want := range defaults {
		if got, err := client.Get(ctx, key).Result(); err != nil || got != want {
			t.Fatalf("%s does not match snapshot: %v", key, err)
		}
		if ttl := server.TTL(key); ttl != 0 {
			t.Fatalf("%s has unexpected TTL: %v", key, ttl)
		}
	}
	if !server.Exists(redisConfigsInitializedKey) || server.TTL(redisConfigsInitializedKey) != 0 {
		t.Fatal("initialization marker must persist")
	}
	if err := client.Set(ctx, "/fg/config/system", "edited", 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := client.Del(ctx, "/fg/config/currency").Err(); err != nil {
		t.Fatal(err)
	}
	inserted, err = initializeRedisConfigs(ctx, client)
	if err != nil || inserted != -1 {
		t.Fatalf("second import must skip: inserted=%d err=%v", inserted, err)
	}
	if client.Get(ctx, "/fg/config/system").Val() != "edited" || server.Exists("/fg/config/currency") {
		t.Fatal("repeat startup overwrote edited or restored deleted configs")
	}
}

func TestInitializeRedisConfigsPreservesExistingData(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	for _, key := range []string{"/fg/config/system", "/fg/config/custom", "player_1"} {
		if err := client.Set(ctx, key, "existing", 0).Err(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := initializeRedisConfigs(ctx, client); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"/fg/config/system", "/fg/config/custom", "player_1"} {
		if got := client.Get(ctx, key).Val(); got != "existing" {
			t.Fatalf("existing key %s changed to %q", key, got)
		}
	}
	if !server.Exists("/fg/config/currency") {
		t.Fatal("missing defaults were not imported")
	}
}

func TestInitializeRedisConfigsConcurrentStarts(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	var wg sync.WaitGroup
	results := make(chan int, 8)
	for i := 0; i < cap(results); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			inserted, err := initializeRedisConfigs(context.Background(), client)
			if err != nil {
				t.Errorf("concurrent import: %v", err)
				return
			}
			results <- inserted
		}()
	}
	wg.Wait()
	close(results)
	imports, skips := 0, 0
	for inserted := range results {
		if inserted == -1 {
			skips++
		} else {
			imports++
		}
	}
	if imports != 1 || skips != 7 {
		t.Fatalf("concurrent starts: imports=%d skips=%d", imports, skips)
	}
}
