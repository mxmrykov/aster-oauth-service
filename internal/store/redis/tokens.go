package redis

import (
	"context"
	"fmt"
	"time"
)

func (r *RedisTc) SetToken(ctx context.Context, signature, token, tp string) error {
	ctx, cancel := context.WithTimeout(ctx, r.MaxPoolInterval)

	defer cancel()

	var dur time.Duration

	if tp == "access" {
		dur = r.AccessTokenDuration
	} else if tp == "refresh" {
		dur = r.RefreshTokenDuration
	}

	return r.Client.Set(ctx, fmt.Sprintf("%s_%s", signature, tp), token, dur).Err()
}

func (r *RedisTc) GetToken(ctx context.Context, signature, tp string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.MaxPoolInterval)

	defer cancel()

	return r.Client.Get(ctx, fmt.Sprintf("%s_%s", signature, tp)).Result()
}

func (r *RedisTc) DeleteToken(ctx context.Context, signature, tp string) error {
	ctx, cancel := context.WithTimeout(ctx, r.MaxPoolInterval)

	defer cancel()

	return r.Client.Del(ctx, fmt.Sprintf("%s_%s", signature, tp)).Err()
}
