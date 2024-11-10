package external_server

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	innerJwt "github.com/mxmrykov/aster-oauth-service/pkg/jwt"
	"github.com/mxmrykov/aster-oauth-service/pkg/responize"
	"net/http"
	"time"
)

func (s *Server) footPrintAuth(gin *gin.Context) {

}

func (s *Server) authorizationMw(gin *gin.Context) {

}

func (s *Server) authenticationMw(ctx *gin.Context) {
	token := ctx.GetHeader("X-TempAuth-Token")
	if token == "" {
		s.svc.Logger().Error().Msg("empty token")
		responize.R(ctx, nil, http.StatusBadRequest, "No X-TempAuth-Token provided", true)
		return
	}

	jwtSecret, err := s.svc.IVault().GetSecret(
		ctx,
		s.svc.OAuth().Vault.TokenRepo.Path,
		s.svc.OAuth().Vault.TokenRepo.AppJwtSecretName,
	)

	if err != nil {
		s.svc.Logger().Error().Err(err).Msg("vault error")
		responize.R(ctx, nil, http.StatusInternalServerError, "Internal authorization error", true)
		return
	}

	payload, err := innerJwt.ValidateXAuthToken(token, jwtSecret)

	if err != nil {
		s.svc.Logger().Error().Err(err).Msg("invalid X-TempAuth-Token")
		responize.R(ctx, nil, http.StatusBadRequest, "Invalid X-TempAuth-Token", true)
		return
	}

	signDate, err := time.Parse(time.RFC3339, payload.SignatureDate)

	if err != nil {
		s.svc.Logger().Error().Err(err).Msg("invalid X-TempAuth-Token")
		responize.R(ctx, nil, http.StatusBadRequest, "Invalid X-TempAuth-Token", true)
		return
	}

	if signDate.Add(5 * time.Minute).Before(time.Now()) {
		s.svc.Logger().Error().Err(err).Msg("X-TempAuth-Token expired")
		responize.R(ctx, nil, http.StatusBadRequest, "X-TempAuth-Token expired", true)
		return
	}

	login, err := s.svc.RedisDc().Get(ctx, payload.Asid)

	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			s.svc.Logger().Error().Err(err).Msg("No such auth or it expired")
			responize.R(ctx, nil, http.StatusUnauthorized, "Unauthorized", true)
			return
		default:
			s.svc.Logger().Error().Err(err).Msg("Redis error")
			responize.R(ctx, nil, http.StatusInternalServerError, "Internal authorization error", true)
			return
		}
	}

	ctx.Set("asid", payload.Asid)
	ctx.Set("login", login)
	ctx.Next()
}
