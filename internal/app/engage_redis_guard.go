package app

import (
	"context"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type engageRedisPublicLeadGuard struct {
	redis *goredis.Client
}

func newEngageRedisPublicLeadGuard(redis *goredis.Client) *engageRedisPublicLeadGuard {
	if redis == nil {
		return nil
	}
	return &engageRedisPublicLeadGuard{redis: redis}
}

func (g *engageRedisPublicLeadGuard) AllowPublicLeadCapture(ctx context.Context, formKey, ipAddress, fingerprint string) (bool, error) {
	key := engagePublicLeadGuardKey(formKey, ipAddress, fingerprint)
	blocked, err := g.redis.Exists(ctx, "engage:lead_capture:block:"+key).Result()
	if err != nil {
		return false, err
	}
	if blocked > 0 {
		return false, nil
	}

	minuteCount, err := g.incrementWithTTL(ctx, "engage:lead_capture:minute:"+key, time.Minute)
	if err != nil {
		return false, err
	}
	hourCount, err := g.incrementWithTTL(ctx, "engage:lead_capture:hour:"+key, time.Hour)
	if err != nil {
		return false, err
	}
	if minuteCount > 5 || hourCount > 30 {
		badCount, err := g.incrementWithTTL(ctx, "engage:lead_capture:bad:"+key, time.Hour)
		if err != nil {
			return false, err
		}
		if badCount >= 3 {
			if err := g.redis.Set(ctx, "engage:lead_capture:block:"+key, "1", 30*time.Minute).Err(); err != nil {
				return false, err
			}
		}
		return false, nil
	}
	return true, nil
}

func (g *engageRedisPublicLeadGuard) RegisterBadPublicLeadCapture(ctx context.Context, formKey, ipAddress string) error {
	key := engagePublicLeadGuardKey(formKey, ipAddress, "")
	badCount, err := g.incrementWithTTL(ctx, "engage:lead_capture:bad:"+key, time.Hour)
	if err != nil {
		return err
	}
	if badCount >= 2 {
		return g.redis.Set(ctx, "engage:lead_capture:block:"+key, "1", 30*time.Minute).Err()
	}
	return nil
}

func (g *engageRedisPublicLeadGuard) incrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	count, err := g.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err := g.redis.Expire(ctx, key, ttl).Err(); err != nil {
			return 0, err
		}
	}
	return count, nil
}

func engagePublicLeadGuardKey(formKey, ipAddress, fingerprint string) string {
	ipAddress = strings.TrimSpace(ipAddress)
	if comma := strings.Index(ipAddress, ","); comma >= 0 {
		ipAddress = strings.TrimSpace(ipAddress[:comma])
	}
	key := formKey + "|ip:" + ipAddress
	if fingerprint != "" {
		key += "|fp:" + strings.TrimSpace(fingerprint)
	}
	return key
}
