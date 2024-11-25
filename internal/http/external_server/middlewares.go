package external_server

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	innerJwt "github.com/mxmrykov/aster-oauth-service/pkg/jwt"
	"github.com/mxmrykov/aster-oauth-service/pkg/responize"
	"net/http"
	"time"
)

func (s *Server) footPrintAuth(gin *gin.Context) {

}

func (s *Server) authorizationMw(ctx *gin.Context) {
	access := ctx.GetHeader("X-Access-Token")

	if access == "" {
		s.svc.Logger().Error().Msg("empty token")
		responize.R(ctx, nil, http.StatusBadRequest, "No X-Access-Token provided", true)
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

	payload, err := innerJwt.ValidateAccessRefreshToken(access, jwtSecret)

	if err != nil {
		v := new(jwt.ValidationError)
		if errors.As(err, &v) {
			if v.Errors&jwt.ValidationErrorExpired != 0 {
				refresh, err := ctx.Request.Cookie("X-Refresh-Token")

				if err != nil {
					s.svc.Logger().Error().Err(err).Msg("invalid X-Refresh-Token")
					responize.R(ctx, nil, http.StatusBadRequest, "Invalid X-Refresh-Token", true)
					return
				}

				refreshPayload, err := innerJwt.ValidateAccessRefreshToken(refresh.Value, jwtSecret)

				if err != nil {
					v := new(jwt.ValidationError)
					if errors.As(err, &v) && v.Errors&jwt.ValidationErrorExpired != 0 {
						s.svc.Logger().Error().Err(err).Msg("Session is expired")
						responize.R(ctx, nil, http.StatusBadRequest, "Session is expired", true)
						return
					}

					s.svc.Logger().Error().Err(err).Msg("Invalid refresh token")
					responize.R(ctx, nil, http.StatusBadRequest, "Unauthorized", true)
					return
				}

				if refreshPayload.Signature != ctx.Request.Header.Get("X-Signature") {
					s.svc.Logger().Error().Msg("Invalid signature")
					responize.R(ctx, nil, http.StatusBadRequest, "Invalid signature", true)
					return
				}

				accessToken, err := s.svc.GenToken(refreshPayload.Iaid, jwtSecret, refreshPayload.Eaid, refreshPayload.Signature, true)

				if err != nil {
					s.svc.Logger().Error().Err(err).Msg("Error generating access token")
					responize.R(ctx, nil, http.StatusInternalServerError, "Internal auth error", true)
					return
				}

				if err := s.svc.RedisTc().SetToken(ctx, refreshPayload.Signature, accessToken, "access"); err != nil {
					s.svc.Logger().Error().Err(err).Msg("Error set access token")
					responize.R(ctx, nil, http.StatusInternalServerError, "Error refresh token", true)
					return
				}

				ctx.Set("refreshed_token", accessToken)
				ctx.Set("iaid", refreshPayload.Iaid)
				ctx.Next()
				return
			}
		} else {
			s.svc.Logger().Error().Err(err).Msg("invalid X-Access-Token")
			responize.R(ctx, nil, http.StatusBadRequest, "Unauthorized", true)
			return
		}

	}

	ctx.Set("signature", payload.Signature)
	ctx.Set("iaid", payload.Iaid)
	ctx.Next()
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
