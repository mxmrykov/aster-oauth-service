package redis

import (
	"context"
	"time"
)

func (r *RedisDc) Set(ctx context.Context, k, v string) error {
	ctx, cancel := context.WithTimeout(ctx, r.MaxPoolInterval)

	defer cancel()

	return r.Client.Set(ctx, k, v, r.oauthExp).Err()
}

func (r *RedisDc) IsAlive(ctx context.Context, k string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.MaxPoolInterval)

	defer cancel()

	ttl, err := r.Client.TTL(ctx, k).Result()

	if err != nil {
		return false, err
	}

	return ttl > 0*time.Second, nil
}

func (r *RedisDc) Get(ctx context.Context, k string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.MaxPoolInterval)

	defer cancel()

	return r.Client.Get(ctx, k).Result()
}
