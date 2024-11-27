package external_server

import (
	"github.com/gin-gonic/gin"
	"github.com/mxmrykov/aster-oauth-service/internal/model"
	"github.com/mxmrykov/aster-oauth-service/pkg/responize"
	"net/http"
)

func (s *Server) exitSession(ctx *gin.Context) {
	iaid, r := ctx.GetString("iaid"), new(model.ExitRequest)

	if iaid == "" {
		s.svc.Logger().Error().Msg("No IAID provided")
		responize.R(ctx, nil, http.StatusUnauthorized, "No IAID provided", true)
		return
	}

	if err := ctx.ShouldBindQuery(r); err != nil {
		s.svc.Logger().Error().Msg("Invalid params")
		responize.R(ctx, nil, http.StatusBadRequest, "Invalid params", true)
		return
	}

	s.svc.Exit(ctx, ctx.Request.Header.Get("X-Signature"), iaid, r.Id)

	responize.R(ctx, nil, http.StatusOK, "Success", true)
}
