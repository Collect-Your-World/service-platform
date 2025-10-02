package redis

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	goRedis "github.com/redis/go-redis/v9"
)

type entry struct {
	value []byte
	expAt *time.Time
}

type InMemoryRedis struct {
	mu    sync.Mutex
	store map[string]entry
}

func NewInMemoryRedis() *InMemoryRedis {
	return &InMemoryRedis{store: make(map[string]entry)}
}

func (m *InMemoryRedis) GetUniversalClient() goRedis.UniversalClient { return nil }
func (m *InMemoryRedis) Reset(ctx context.Context) error {
	m.mu.Lock()
	m.store = make(map[string]entry)
	m.mu.Unlock()
	return nil
}
func (m *InMemoryRedis) Close() error { return nil }

func (m *InMemoryRedis) SetPrimitive(ctx context.Context, key string, value any, ttl time.Duration) error {
	b, _ := json.Marshal(value)
	return m.Set(ctx, key, b, ttl)
}

func (m *InMemoryRedis) GetPrimitive(ctx context.Context, key string, outPtr any) error {
	var b []byte
	if err := m.getBytes(ctx, key, &b); err != nil {
		return err
	}
	return json.Unmarshal(b, outPtr)
}

func (m *InMemoryRedis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	default:
		bb, err := json.Marshal(v)
		if err != nil {
			return err
		}
		b = bb
	}
	var exp *time.Time
	if ttl > 0 {
		t := time.Now().Add(ttl)
		exp = &t
	}
	m.store[key] = entry{value: b, expAt: exp}
	return nil
}

func (m *InMemoryRedis) Get(ctx context.Context, key string, outPtr any) error {
	var b []byte
	if err := m.getBytes(ctx, key, &b); err != nil {
		return err
	}
	return json.Unmarshal(b, outPtr)
}

func (m *InMemoryRedis) getBytes(_ context.Context, key string, out *[]byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.store[key]
	if !ok {
		return goRedis.Nil
	}
	if e.expAt != nil && time.Now().After(*e.expAt) {
		delete(m.store, key)
		return goRedis.Nil
	}
	*out = e.value
	return nil
}

func (m *InMemoryRedis) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	delete(m.store, key)
	m.mu.Unlock()
	return nil
}
func (m *InMemoryRedis) Increment(ctx context.Context, key string, val int64) error { return nil }
func (m *InMemoryRedis) Decrement(ctx context.Context, key string, val int64) error { return nil }
func (m *InMemoryRedis) SetExpire(ctx context.Context, key string, ttl time.Duration) error {
	return m.Expire(ctx, key, ttl)
}
func (m *InMemoryRedis) AcquireLock(ctx context.Context, lockKey string, ttl time.Duration) (bool, error) {
	return true, nil
}
func (m *InMemoryRedis) ReleaseLock(ctx context.Context, lockKey string) error { return nil }
func (m *InMemoryRedis) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.store[key]
	return ok, nil
}
func (m *InMemoryRedis) Scan(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return nil, 0, nil
}

func (m *InMemoryRedis) HSet(ctx context.Context, key string, fields map[string]interface{}) error {
	return nil
}
func (m *InMemoryRedis) HDel(ctx context.Context, key string, field string) error { return nil }
func (m *InMemoryRedis) Expire(ctx context.Context, key string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.store[key]
	if !ok {
		return errors.New("not found")
	}
	if ttl > 0 {
		t := time.Now().Add(ttl)
		e.expAt = &t
	} else {
		e.expAt = nil
	}
	m.store[key] = e
	return nil
}
func (m *InMemoryRedis) HVals(ctx context.Context, key string) ([]string, error) { return nil, nil }
func (m *InMemoryRedis) HExists(ctx context.Context, key, field string) (bool, error) {
	return false, nil
}
func (m *InMemoryRedis) HLen(ctx context.Context, key string) (int64, error)         { return 0, nil }
func (m *InMemoryRedis) HGet(ctx context.Context, key, field string) (string, error) { return "", nil }
func (m *InMemoryRedis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return map[string]string{}, nil
}

func (m *InMemoryRedis) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return nil
}
func (m *InMemoryRedis) SMembers(ctx context.Context, key string) ([]string, error) {
	return []string{}, nil
}
func (m *InMemoryRedis) SRem(ctx context.Context, key string, members ...any) error { return nil }
func (m *InMemoryRedis) SCard(ctx context.Context, key string) (int64, error)       { return 0, nil }

func (m *InMemoryRedis) TxPipeline(ctx context.Context, fn func(pipe goRedis.Pipeliner) error) ([]goRedis.Cmder, error) {
	return nil, nil
}
func (m *InMemoryRedis) GetAllKeysWithPattern(ctx context.Context, pattern string) ([]string, error) {
	return []string{}, nil
}
