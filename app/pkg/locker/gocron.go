package locker

import (
	"time"

	redislock "github.com/go-co-op/gocron-redis-lock/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
)

type Locker gocron.Locker

func NewLocker(rd redis.UniversalClient) (Locker, error) {
	return redislock.NewRedisLockerAlways(rd, redislock.WithExpiry(1*time.Minute))
}

func NewTryLocker(rd redis.UniversalClient) (Locker, error) {
	return redislock.NewRedisLockerAlways(rd, redislock.WithExpiry(1*time.Minute), redislock.WithTries(1))
}
