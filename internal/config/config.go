package config

import (
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/mxmrykov/aster-oauth-service/pkg/logger"
	"github.com/rs/zerolog"
	"os"
	"time"
)

type (
	OAuth struct {
		UseStackTrace bool `yaml:"useStackTrace"`

		DcRedis        DcRedis        `yaml:"dcRedis"`
		TcRedis        TcRedis        `yaml:"tcRedis"`
		ClientPostgres ClientPostgres `yaml:"clientPostgres"`
		UserPostgres   UserPostgres   `yaml:"userPostgres"`

		Vault          Vault          `yaml:"vault"`
		GrpcServer     GrpcServer     `yaml:"grpcServer"`
		ExternalServer ExternalServer `yaml:"externalServer"`
	}

	DcRedis struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`

		MaxPoolInterval time.Duration `yaml:"maxPoolInterval"`
		OAuthCodeExp    time.Duration `yaml:"oauthCodeExp"`
		ConfirmCodeExp  time.Duration `yaml:"confirmCodeExp"`
	}

	TcRedis struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`

		MaxPoolInterval      time.Duration `yaml:"maxPoolInterval"`
		AccessTokenDuration  time.Duration `yaml:"accessTokenDuration"`
		RefreshTokenDuration time.Duration `yaml:"refreshTokenDuration"`
	}

	ClientPostgres struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		DataBaseName string `yaml:"dataBaseName"`

		MaxPoolInterval time.Duration `yaml:"maxPoolInterval"`
	}

	UserPostgres struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		DataBaseName string `yaml:"dataBaseName"`

		MaxPoolInterval time.Duration `yaml:"maxPoolInterval"`
	}

	Vault struct {
		AuthToken     string        `env:"VAULT_AUTH_TOKEN"`
		Host          string        `yaml:"host"`
		Port          int           `yaml:"port"`
		ClientTimeout time.Duration `yaml:"clientTimeout"`

		TokenRepo struct {
			Path string `yaml:"path"`

			OAuthJwtSecretName string `yaml:"oauthJwtSecretName"`
			AppJwtSecretName   string `yaml:"appJwtSecretName"`
		} `yaml:"tokenRepo"`

		RedisSecret struct {
			Path string `yaml:"path"`

			DcRedisSecretName string `yaml:"dcRedisSecretName"`
			DcRedisUserName   string `yaml:"dcRedisUserName"`
			TcRedisSecretName string `yaml:"tcRedisSecretName"`
			TcRedisUserName   string `yaml:"tcRedisUserName"`
		} `yaml:"redisSecret"`

		PostgresSecret struct {
			Path string `yaml:"path"`

			ClPostgresSecretName string `yaml:"clPostgresSecretName"`
			ClPostgresUserName   string `yaml:"clPostgresUserName"`
			UPostgresSecretName  string `yaml:"uPostgresSecretName"`
			UPostgresUserName    string `yaml:"uPostgresUserName"`
		} `yaml:"postgresSecret"`
	}

	GrpcServer struct {
		Port        int           `yaml:"port"`
		MaxPollTime time.Duration `yaml:"maxPollTime"`
	}

	ExternalServer struct {
		Port                    int           `yaml:"port"`
		RateLimiterTimeframe    time.Duration `yaml:"rateLimiterTimeframe"`
		RateLimiterCap          uint8         `yaml:"rateLimiterCap"`
		RateLimitCookieLifetime int           `yaml:"rateLimitCookieLifetime"`
	}
)

func InitConfig() (*OAuth, *zerolog.Logger, error) {
	cfg := *new(OAuth)

	if os.Getenv("BUILD_ENV") == "" {
		return nil, nil, errors.New("build environment is not assigned")
	}

	path := fmt.Sprintf("./deploy/%s.yaml", os.Getenv("BUILD_ENV"))

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, nil, err
	}

	l := logger.NewLogger(cfg.UseStackTrace)

	return &cfg, l, nil
}
