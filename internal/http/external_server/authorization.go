package external_server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mxmrykov/aster-oauth-service/pkg/responize"
)

func (s *Server) exitSession(ctx *gin.Context) {
	iaid := ctx.GetString("iaid")

	if iaid == "" {
		s.svc.Logger().Error().Msg("No IAID provided")
		responize.R(ctx, nil, http.StatusUnauthorized, "No IAID provided", true)
		return
	}

	s.svc.Exit(ctx, ctx.Request.Header.Get("X-Signature"), iaid)

	responize.R(ctx, nil, http.StatusOK, "Success", true)
}
