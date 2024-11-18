package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/mxmrykov/aster-oauth-service/pkg/utils"
)

func (s *Service) SetPhoneConfirmCode(ctx *gin.Context, phone string) error {
	code := utils.GetConfirmCode()
	return s.IRedisDc.SetConfirmCode(ctx, phone, code)
}

func (s *Service) IfPhoneInUse(ctx *gin.Context, phone string) (bool, error) {
	return s.IUserStore.IsPhoneInUse(ctx, phone)
}

func (s *Service) IfLoginInUse(ctx *gin.Context, login string) (bool, error) {
	return s.IUserStore.IsLoginInUse(ctx, login)
}

func (s *Service) GetPhoneConfirmCode(ctx *gin.Context, phone string) (string, error) {
	return s.IRedisDc.Get(ctx, phone)
}

func (s *Service) IfCodeSent(ctx *gin.Context, phone string) (bool, error) {
	_, err := s.IRedisDc.Get(ctx, phone)

	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
