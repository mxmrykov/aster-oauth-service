package redis

import (
	"context"
	"time"
)

func (r *RedisDc) SetOAuthCode(ctx context.Context, k, v string) error {
	return r.set(ctx, k, v, r.oauthExp)
}

func (r *RedisDc) SetConfirmCode(ctx context.Context, k, v string) error {
	return r.set(ctx, k, v, r.confirmCodeExp)
}

func (r *RedisDc) Get(ctx context.Context, k string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.MaxPoolInterval)

	defer cancel()

	return r.Client.Get(ctx, k).Result()
}

func (r *RedisDc) set(ctx context.Context, k string, v interface{}, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, r.MaxPoolInterval)

	defer cancel()

	return r.Client.Set(ctx, k, v, ttl).Err()
}
