package service

import "github.com/gin-gonic/gin"

func (s *Service) Exit(ctx *gin.Context, signature, iaid string) {
	_ = s.RedisTc().DeleteToken(ctx, signature, "refresh")
	_ = s.RedisTc().DeleteToken(ctx, signature, "access")

	_ = s.UserStore().Exit(ctx, iaid, signature)
}
