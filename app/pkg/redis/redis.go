package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"backend/service-platform/app/internal/config"
)

type Redis interface {
	GetUniversalClient() redis.UniversalClient
	Reset(ctx context.Context) error
	Close() error
	// Optimal method to use to set a value into redis. It supports only primitive types like string, int, float, etc.
	// The detail can be checked at https://github.com/redis/go-redis/blob/master/internal/proto/scan.go
	SetPrimitive(ctx context.Context, key string, value any, ttl_sec time.Duration) error

	// Optimal method to use to get a value into redis. It supports only primitive types like string, int, float, etc.
	// The detail can be checked at https://github.com/redis/go-redis/blob/master/internal/proto/scan.go
	GetPrimitive(ctx context.Context, key string, outPtr any) error

	// Default method to use to set a value into redis. It supports any type of value including structs, maps, slices, etc.
	Set(ctx context.Context, key string, value any, ttl_sec time.Duration) error

	// Default method to use to get a value into redis. It supports any type of value including structs, maps, slices, etc.
	Get(ctx context.Context, key string, outPtr any) error
	Delete(ctx context.Context, key string) error
	Increment(ctx context.Context, key string, val int64) error
	Decrement(ctx context.Context, key string, val int64) error
	SetExpire(ctx context.Context, key string, ttl_sec time.Duration) error
	AcquireLock(ctx context.Context, lockKey string, ttl_sec time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, lockKey string) error
	Exists(ctx context.Context, key string) (bool, error)
	Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error)

	HSet(ctx context.Context, key string, fields map[string]interface{}) error
	HDel(ctx context.Context, key string, field string) error
	Expire(ctx context.Context, key string, ttl time.Duration) error
	HVals(ctx context.Context, key string) ([]string, error)
	HExists(ctx context.Context, key, field string) (bool, error)
	HLen(ctx context.Context, key string) (int64, error)
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)

	SAdd(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	SRem(ctx context.Context, key string, members ...any) error
	SCard(ctx context.Context, key string) (int64, error)

	TxPipeline(ctx context.Context, fn func(pipe redis.Pipeliner) error) ([]redis.Cmder, error)
	GetAllKeysWithPattern(
		ctx context.Context,
		pattern string,
	) ([]string, error)
}

type Client struct {
	client *redis.ClusterClient
	log    *zap.Logger
}

func NewRedisClusterClient(cfg config.RedisConfig, log *zap.Logger) (Redis, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           strings.Split(cfg.Hosts, ","),
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
		ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
		ConnMaxLifetime: time.Duration(cfg.ConnMaxLifetime) * time.Second,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Client{
		client: client,
		log:    log,
	}, nil
}

func (r *Client) GetUniversalClient() redis.UniversalClient {
	return r.client
}

func (r *Client) Close() error {
	return r.client.Close()
}

func (r *Client) Increment(ctx context.Context, key string, val int64) error {
	return r.client.IncrBy(ctx, key, val).Err()
}

func (r *Client) Decrement(ctx context.Context, key string, val int64) error {
	return r.client.DecrBy(ctx, key, val).Err()
}

func (r *Client) SetExpire(ctx context.Context, key string, dur time.Duration) error {
	return r.client.Expire(ctx, key, dur).Err()
}

func (r *Client) SetPrimitive(c context.Context, key string, value any, ttls time.Duration) error {
	return r.client.Set(c, key, value, ttls).Err()
}

func (r *Client) GetPrimitive(c context.Context, key string, outPtr any) error {
	return r.client.Get(c, key).Scan(outPtr)
}

func (r *Client) Set(c context.Context, key string, value any, ttls time.Duration) error {
	b, err := json.Marshal(value)
	if err == nil {
		return r.client.Set(c, key, b, ttls).Err()
	}
	return err
}

func (r *Client) Get(c context.Context, key string, outPtr any) error {
	b, err := r.client.Get(c, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, outPtr)
}

func (r *Client) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	result := r.client.Scan(ctx, cursor, match, count)
	if err := result.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to scan redis: %w", err)
	}
	keys, nextCursor := result.Val()
	return keys, nextCursor, nil
}

func (r *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

func (r *Client) SRem(ctx context.Context, key string, members ...any) error {
	return r.client.SRem(ctx, key, members...).Err()
}

func (r *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	result := r.client.SMembers(ctx, key)
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan redis: %w", err)
	}
	return result.Val(), nil
}

func (r *Client) SCard(ctx context.Context, key string) (int64, error) {
	result := r.client.SCard(ctx, key)
	if err := result.Err(); err != nil {
		return 0, fmt.Errorf("failed to get set cardinality: %w", err)
	}
	return result.Val(), nil
}

func Wrap[T any](
	c context.Context,
	r Redis,
	key string,
	model *T,
	ttls time.Duration,
	callback func() (T, error),
) (err error) {
	if err = r.Get(c, key, model); err != nil {
		res, err := callback()
		if nil != err {
			return err
		}
		*model = res
		return r.Set(c, key, res, ttls)
	}
	return err
}

func (r *Client) AcquireLock(ctx context.Context, lockKey string, lockDuration time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, lockKey, true, lockDuration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return ok, nil
}

func (r *Client) ReleaseLock(ctx context.Context, lockKey string) error {
	return r.Delete(ctx, lockKey)
}

func (r *Client) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, SkipNotFound(err)
	}
	return result >= 1, nil
}

func SkipNotFound(err error) error {
	if errors.Is(err, redis.Nil) {
		return nil
	}
	return err
}

func (r *Client) HSet(ctx context.Context, key string, fields map[string]any) error {
	return r.client.HSet(ctx, key, fields).Err()
}

func (r *Client) HDel(ctx context.Context, key string, field string) error {
	return r.client.HDel(ctx, key, field).Err()
}

func (r *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

func (r *Client) HVals(ctx context.Context, key string) ([]string, error) {
	return r.client.HVals(ctx, key).Result()
}

func (r *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

func (r *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

func (r *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	return r.client.HExists(ctx, key, field).Result()
}

func (r *Client) HLen(ctx context.Context, key string) (int64, error) {
	return r.client.HLen(ctx, key).Result()
}

func (r *Client) Reset(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

func (r *Client) TxPipeline(ctx context.Context, fn func(pipe redis.Pipeliner) error) ([]redis.Cmder, error) {
	transaction := r.client.TxPipeline()
	if err := fn(transaction); err != nil {
		return nil, fmt.Errorf("failed to execute transaction: %w", err)
	}
	cmds, err := transaction.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transaction commands: %w", err)
	}
	return cmds, nil
}

func (r *Client) GetAllKeysWithPattern(
	ctx context.Context,
	pattern string,
) ([]string, error) {
	keys := []string{}
	err := r.client.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
		eachMasterKeys, err := client.Keys(ctx, fmt.Sprintf("%s*", pattern)).Result()
		if err != nil {
			return err
		}
		keys = append(keys, eachMasterKeys...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}
