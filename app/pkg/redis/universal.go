package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"backend/service-platform/app/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// UniversalClient wraps redis.UniversalClient to handle both single-node and cluster setups
type UniversalClient struct {
	client    redis.UniversalClient
	log       *zap.Logger
	isCluster bool
}

// NewUniversalRedisClient creates a Redis client that can handle both single-node and cluster configurations
func NewUniversalRedisClient(cfg config.RedisConfig, log *zap.Logger) (Redis, error) {
	hosts := strings.Split(cfg.Hosts, ",")

	// Trim whitespace from hosts
	for i, host := range hosts {
		hosts[i] = strings.TrimSpace(host)
	}

	var client redis.UniversalClient
	var isCluster bool

	// If only one host is provided, use single-node client
	if len(hosts) == 1 {
		client = redis.NewClient(&redis.Options{
			Addr:            hosts[0],
			PoolSize:        cfg.PoolSize,
			MinIdleConns:    cfg.MinIdleConns,
			MaxIdleConns:    cfg.MaxIdleConns,
			WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
			ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
			ConnMaxLifetime: time.Duration(cfg.ConnMaxLifetime) * time.Second,
		})
		isCluster = false
	} else {
		// Multiple hosts - try cluster first, fallback to failover
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:           hosts,
			PoolSize:        cfg.PoolSize,
			MinIdleConns:    cfg.MinIdleConns,
			MaxIdleConns:    cfg.MaxIdleConns,
			WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
			ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
			ConnMaxLifetime: time.Duration(cfg.ConnMaxLifetime) * time.Second,
		})
		isCluster = true
	}

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		if isCluster {
			// Cluster failed, try failover client instead
			log.Warn("Redis cluster connection failed, trying failover client", zap.Error(err))
			client.Close()

			client = redis.NewFailoverClient(&redis.FailoverOptions{
				MasterName:      "mymaster", // Default sentinel master name
				SentinelAddrs:   hosts,
				PoolSize:        cfg.PoolSize,
				MinIdleConns:    cfg.MinIdleConns,
				MaxIdleConns:    cfg.MaxIdleConns,
				WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
				ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
				ConnMaxLifetime: time.Duration(cfg.ConnMaxLifetime) * time.Second,
			})
			isCluster = false

			if err := client.Ping(context.Background()).Err(); err != nil {
				return nil, fmt.Errorf("failed to ping redis with cluster or failover: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to ping redis: %w", err)
		}
	}

	log.Info("Redis client connected",
		zap.Strings("hosts", hosts),
		zap.Bool("is_cluster", isCluster),
		zap.String("client_type", getClientType(client)))

	return &UniversalClient{
		client:    client,
		log:       log,
		isCluster: isCluster,
	}, nil
}

func getClientType(client redis.UniversalClient) string {
	switch client.(type) {
	case *redis.Client:
		return "single-node"
	case *redis.ClusterClient:
		return "cluster"
	case *redis.Ring:
		return "ring"
	default:
		return "unknown"
	}
}

func (r *UniversalClient) GetUniversalClient() redis.UniversalClient {
	return r.client
}

func (r *UniversalClient) Close() error {
	return r.client.Close()
}

func (r *UniversalClient) Increment(ctx context.Context, key string, val int64) error {
	return r.client.IncrBy(ctx, key, val).Err()
}

func (r *UniversalClient) Decrement(ctx context.Context, key string, val int64) error {
	return r.client.DecrBy(ctx, key, val).Err()
}

func (r *UniversalClient) SetExpire(ctx context.Context, key string, dur time.Duration) error {
	return r.client.Expire(ctx, key, dur).Err()
}

func (r *UniversalClient) SetPrimitive(c context.Context, key string, value any, ttls time.Duration) error {
	return r.client.Set(c, key, value, ttls).Err()
}

func (r *UniversalClient) GetPrimitive(c context.Context, key string, outPtr any) error {
	return r.client.Get(c, key).Scan(outPtr)
}

func (r *UniversalClient) Set(c context.Context, key string, value any, ttls time.Duration) error {
	b, err := json.Marshal(value)
	if err == nil {
		return r.client.Set(c, key, b, ttls).Err()
	}
	return err
}

func (r *UniversalClient) Get(c context.Context, key string, outPtr any) error {
	b, err := r.client.Get(c, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, outPtr)
}

func (r *UniversalClient) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	result := r.client.Scan(ctx, cursor, match, count)
	if err := result.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to scan redis: %w", err)
	}
	keys, nextCursor := result.Val()
	return keys, nextCursor, nil
}

func (r *UniversalClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

func (r *UniversalClient) SRem(ctx context.Context, key string, members ...any) error {
	return r.client.SRem(ctx, key, members...).Err()
}

func (r *UniversalClient) SMembers(ctx context.Context, key string) ([]string, error) {
	result := r.client.SMembers(ctx, key)
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("failed to get set members: %w", err)
	}
	return result.Val(), nil
}

func (r *UniversalClient) SCard(ctx context.Context, key string) (int64, error) {
	result := r.client.SCard(ctx, key)
	if err := result.Err(); err != nil {
		return 0, fmt.Errorf("failed to get set cardinality: %w", err)
	}
	return result.Val(), nil
}

func (r *UniversalClient) AcquireLock(ctx context.Context, lockKey string, lockDuration time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, lockKey, true, lockDuration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return ok, nil
}

func (r *UniversalClient) ReleaseLock(ctx context.Context, lockKey string) error {
	return r.Delete(ctx, lockKey)
}

func (r *UniversalClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *UniversalClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, SkipNotFound(err)
	}
	return result >= 1, nil
}

func (r *UniversalClient) HSet(ctx context.Context, key string, fields map[string]any) error {
	return r.client.HSet(ctx, key, fields).Err()
}

func (r *UniversalClient) HDel(ctx context.Context, key string, field string) error {
	return r.client.HDel(ctx, key, field).Err()
}

func (r *UniversalClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

func (r *UniversalClient) HVals(ctx context.Context, key string) ([]string, error) {
	return r.client.HVals(ctx, key).Result()
}

func (r *UniversalClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

func (r *UniversalClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

func (r *UniversalClient) HExists(ctx context.Context, key, field string) (bool, error) {
	return r.client.HExists(ctx, key, field).Result()
}

func (r *UniversalClient) HLen(ctx context.Context, key string) (int64, error) {
	return r.client.HLen(ctx, key).Result()
}

func (r *UniversalClient) Reset(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

func (r *UniversalClient) TxPipeline(ctx context.Context, fn func(pipe redis.Pipeliner) error) ([]redis.Cmder, error) {
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

func (r *UniversalClient) GetAllKeysWithPattern(
	ctx context.Context,
	pattern string,
) ([]string, error) {
	keys := []string{}

	if r.isCluster {
		// For cluster mode, iterate through all masters
		if clusterClient, ok := r.client.(*redis.ClusterClient); ok {
			err := clusterClient.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
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
		} else {
			// Fallback for non-cluster universal clients
			allKeys, err := r.client.Keys(ctx, fmt.Sprintf("%s*", pattern)).Result()
			if err != nil {
				return nil, err
			}
			keys = allKeys
		}
	} else {
		// For single-node mode, use direct keys command
		allKeys, err := r.client.Keys(ctx, fmt.Sprintf("%s*", pattern)).Result()
		if err != nil {
			return nil, err
		}
		keys = allKeys
	}

	return keys, nil
}
