package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/mxmrykov/aster-oauth-service/internal/config"
	"time"
)

type IRedisTc interface {
	SetToken(ctx context.Context, signature, token, tp string) error
	GetToken(ctx context.Context, signature, tp string) (string, error)
	IsSignatureAlive(ctx context.Context, signature, tp string) (bool, error)
}

type RedisTc struct {
	Client               *redis.Client
	MaxPoolInterval      time.Duration
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

func NewRedisTc(cfg *config.TcRedis, user, password string) IRedisTc {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Username: user,
		Password: password,
		DB:       2,
	})

	return &RedisTc{
		Client:               rdb,
		MaxPoolInterval:      cfg.MaxPoolInterval,
		AccessTokenDuration:  cfg.AccessTokenDuration,
		RefreshTokenDuration: cfg.RefreshTokenDuration,
	}
}
