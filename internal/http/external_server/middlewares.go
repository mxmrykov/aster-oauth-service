package external_server

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/mxmrykov/aster-oauth-service/internal/cache"
	jwt2 "github.com/mxmrykov/aster-oauth-service/pkg/jwt"
	"github.com/mxmrykov/aster-oauth-service/pkg/responize"
)

func (s *Server) footPrintAuth(ctx *gin.Context) {
	c, err := ctx.Request.Cookie("X-Client-Footprint")

	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			ctx.Set("X-Client-Footprint", s.setFootprintCookie(ctx))
		default:
			responize.R(ctx, nil, http.StatusBadRequest, "Footprint: signature failed", true)
			return
		}
	} else {
		fpClient := s.svc.ICache().GetClient(c.Value)

		switch {
		case c.Value == "":
			ctx.Set("X-Client-Footprint", s.setFootprintCookie(ctx))
			ctx.Next()
		case fpClient == nil:
			s.dropFootprintCookie(ctx)

			responize.R(ctx, nil, http.StatusUnauthorized, "Footprint: invalid signature", true)
			return
		case fpClient.RateLimitRemain <= 1:
			if !time.Now().After(fpClient.LastReq.Add(s.svc.OAuth().ExternalServer.RateLimiterTimeframe)) {
				fpClient.LastReq = time.Now()
				fpClient.LastUpdated = time.Now()
				s.svc.ICache().SetClient(c.Value, fpClient)

				responize.R(ctx, nil, http.StatusTooManyRequests, "Footprint: rate limited", true)
				return
			}

			fpClient.RateLimitRemain = s.svc.OAuth().ExternalServer.RateLimiterCap
		default:
			fpClient.RateLimitRemain -= 1
		}

		fpClient.LastReq = time.Now()
		fpClient.LastUpdated = time.Now()

		s.svc.ICache().SetClient(c.Value, fpClient)
		ctx.Set("X-Client-Footprint", c.Value)
	}

	ctx.Next()
}

func (s *Server) internalAuthMiddleWare(ctx *gin.Context) {
	authToken := ctx.GetHeader("X-Auth-Token")

	if authToken == "" {
		s.svc.Logger().Error().Msg("Empty auth token")
		responize.R(ctx, nil, http.StatusBadRequest, "Empty auth token", true)
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

	authPayload, err := jwt2.ValidateAsidToken(authToken, jwtSecret)

	if err != nil {
		s.svc.Logger().Error().Err(err).Msg("token error")
		responize.R(ctx, nil, http.StatusBadRequest, "Invalid token", true)
		return
	}

	ctx.Set("iaid", authPayload.Iaid)
	ctx.Next()
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

	payload, err := jwt2.ValidateAccessRefreshToken(access, jwtSecret)

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

				refreshPayload, err := jwt2.ValidateAccessRefreshToken(refresh.Value, jwtSecret)

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
			} else {
				s.svc.Logger().Error().Err(err).Msg("invalid X-Access-Token")
				responize.R(ctx, nil, http.StatusBadRequest, "Invalid X-Access-Token", true)
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

	payload, err := jwt2.ValidateXAuthToken(token, jwtSecret)

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

func (s *Server) dropFootprintCookie(ctx *gin.Context) {
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "X-Client-Footprint",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) setFootprintCookie(ctx *gin.Context) string {
	sign := base64.StdEncoding.EncodeToString(
		[]byte(
			uuid.New().String(),
		),
	)

	s.svc.ICache().SetClient(strings.ToUpper(sign), &cache.Props{
		RateLimitRemain: 5,
		LastReq:         time.Now(),
		LastUpdated:     time.Now(),
	})

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "X-Client-Footprint",
		Value:    strings.ToUpper(sign),
		Path:     "/",
		MaxAge:   s.svc.OAuth().ExternalServer.RateLimitCookieLifetime,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	return strings.ToUpper(sign)
}
