package service

import "github.com/gin-gonic/gin"

func (s *Service) SetPhoneConfirmCode(ctx *gin.Context, phone string) error {
	// TODO: realization of sms-code
	code := "211004"
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
