package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/mxmrykov/aster-oauth-service/internal/config"
	"time"
)

var ErrorNotFound = errors.New("not found")

type IRedisDc interface {
	SetConfirmCode(ctx context.Context, k, v string) error
	SetOAuthCode(ctx context.Context, k, v string) error
	IsAlive(ctx context.Context, k string) (bool, error)
	Get(ctx context.Context, k string) (string, error)
}

type RedisDc struct {
	Client          *redis.Client
	MaxPoolInterval time.Duration
	oauthExp        time.Duration
	confirmCodeExp  time.Duration
}

func NewRedisDc(cfg *config.DcRedis, user, password string) IRedisDc {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Username: user,
		Password: password,
		DB:       0,
	})

	return &RedisDc{
		Client:          rdb,
		MaxPoolInterval: cfg.MaxPoolInterval,
		oauthExp:        cfg.OAuthCodeExp,
	}
}
