package redis

import (
	"context"
	"fmt"
)

func (r *RedisTc) SetToken(ctx context.Context, signature, token, tp string) error {
	ctx, cancel := context.WithTimeout(ctx, r.MaxPoolInterval)

	defer cancel()

	return r.Client.Set(ctx, fmt.Sprintf("%s_%s", signature, tp), token, r.AccessTokenDuration).Err()
}

func (r *RedisTc) GetToken(ctx context.Context, signature, tp string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.MaxPoolInterval)

	defer cancel()

	return r.Client.Get(ctx, fmt.Sprintf("%s_%s", signature, tp)).Result()
}
