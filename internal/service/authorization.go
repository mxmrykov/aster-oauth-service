package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/mxmrykov/aster-oauth-service/internal/model"
	"github.com/mxmrykov/aster-oauth-service/pkg/utils"
)

func (s *Service) ValidateClientAuth(ctx context.Context, r *model.AuthRequest, iaid string) error {
	if r.OAuthCode == "" || r.ClientID == "" || r.ClientSecret == "" {
		return errors.New("missing some data")
	}

	v, err := s.IRedisDc.Get(ctx, r.OAuthCode)

	if err != nil && !errors.Is(err, redis.Nil) {
		return errors.New("cannot confirm code")
	}

	if v != iaid {
		return errors.New("wrong confirm code")
	}

	if err = s.ClientStore().CheckClient(ctx, r.ClientID, r.ClientSecret, iaid); err != nil {
		return errors.New("wrong client data")
	}

	return nil
}

func (s *Service) ResourceOwnerAuthorize(ctx *gin.Context, iaid string) (*model.AuthDTO, error) {
	oauthSecret, err := s.Vault.GetSecret(ctx, s.Cfg.Vault.TokenRepo.Path, s.Cfg.Vault.TokenRepo.OAuthJwtSecretName)

	if err != nil {
		return nil, err
	}

	eid, login, err := s.UserStore().ExtractEaid(ctx, iaid)

	if err != nil {
		return nil, err
	}

	signature := strings.ToUpper(uuid.New().String())

	accessToken, err := s.GenToken(iaid, oauthSecret, strconv.Itoa(eid), signature, true)

	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenToken(iaid, oauthSecret, signature, strconv.Itoa(eid))

	if err != nil {
		return nil, err
	}

	device := utils.GetDeviceInfo(ctx.Request.UserAgent())

	if err = s.IUserStore.EnterSession(ctx, model.EnterSession{
		Iaid: iaid,
		InternalSignUpRequest: model.InternalSignUpRequest{
			Ip:             ctx.Request.RemoteAddr,
			DeviceName:     fmt.Sprintf("%s %s, %s", device.OSName, device.OSVersion, device.DeviceName),
			DevicePlatform: fmt.Sprintf("%s, %s", device.Client, device.ClientVersion),
		},
	}); err != nil {
		return nil, err
	}

	if err = s.RedisTc().SetToken(ctx, signature, accessToken, "access"); err != nil {
		return nil, err
	}

	if err = s.RedisTc().SetToken(ctx, signature, refreshToken, "refresh"); err != nil {
		return nil, err
	}

	if err = s.RedisDc().SetIAID(ctx, login, iaid); err != nil {
		return nil, err
	}

	return &model.AuthDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Signature:    signature,
	}, nil
}
