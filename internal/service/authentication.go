package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/mxmrykov/aster-oauth-service/internal/model"
	"github.com/mxmrykov/aster-oauth-service/pkg/hashing"
	"github.com/mxmrykov/aster-oauth-service/pkg/jwt"
	"github.com/mxmrykov/aster-oauth-service/pkg/utils"
	"strconv"
	"strings"
	"time"
)

func (s *Service) SetPhoneConfirmCode(ctx *gin.Context, phone string) error {
	code := utils.GetConfirmCode()
	return s.IRedisDc.SetConfirmCode(ctx, phone, strconv.Itoa(code))
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

func (s *Service) SetPhoneConfirmed(ctx *gin.Context, phone string) error {
	return s.IRedisDc.SetConfirmCode(ctx, phone, "APPROVED")
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

func (s *Service) ValidateUserSignup(ctx *gin.Context, r *model.SignupRequest) error {
	asid, mwLogin := ctx.GetString("asid"), ctx.GetString("login")
	ok, err := s.approveSignupAsid(ctx, asid, mwLogin)

	if err != nil {
		return err
	}

	a, err := s.IRedisDc.Get(ctx, r.Phone)

	switch {
	case err != nil:
		return err
	case a != "APPROVED":
		return errors.New("phone is not confirmed")
	case errors.Is(err, redis.Nil):
		return errors.New("no such phone to confirm")
	}

	if mwLogin != r.Login || !ok {
		return errors.New("incorrect token login")
	}

	e, err := s.IfPhoneInUse(ctx, r.Phone)

	switch {
	case err != nil:
		return errors.New("failed to check if phone is in use")
	case e:
		return errors.New("phone already in use")
	}

	e, err = s.IfLoginInUse(ctx, r.Phone)

	switch {
	case err != nil:
		return errors.New("failed to check if login is in use")
	case e:
		return errors.New("login already in use")
	}

	return nil
}

func (s *Service) SignupUser(ctx *gin.Context, r *model.SignupRequest) (*model.SignUpDTO, error) {
	oauthSecret, err := s.Vault.GetSecret(ctx, s.Cfg.Vault.TokenRepo.Path, s.Cfg.Vault.TokenRepo.OAuthJwtSecretName)

	if err != nil {
		return nil, err
	}

	Eaid, Iaid, signature :=
		strconv.FormatInt(time.Now().Unix(), 10),
		base64.StdEncoding.EncodeToString([]byte(uuid.New().String())),
		strings.ToUpper(uuid.New().String())

	accessToken, err := s.GenToken(Iaid, oauthSecret, Eaid, signature, true)

	if err != nil {
		return nil, err
	}

	refreshToken, err := s.GenToken(Iaid, oauthSecret, signature, Eaid)

	if err != nil {
		return nil, err
	}

	utx, err := s.IUserStore.BeginTx(ctx)

	if err != nil {
		return nil, err
	}

	cltx, err := s.IClientStore.BeginTx(ctx)

	if err != nil {
		return nil, err
	}

	device := utils.GetDeviceInfo(ctx.Request.UserAgent())

	pwd, _ := hashing.New(r.Password)

	if err = s.IUserStore.SignUpUser(ctx, utx,
		model.ExternalSignUpRequest{
			Iaid:     Iaid,
			Eaid:     Eaid,
			Name:     r.Name,
			Login:    r.Login,
			Phone:    r.Phone,
			Password: pwd,
		},
		model.InternalSignUpRequest{
			Ip:             ctx.Request.RemoteAddr,
			DeviceName:     fmt.Sprintf("%s %s, %s", device.OSName, device.OSVersion, device.DeviceName),
			DevicePlatform: fmt.Sprintf("%s, %s", device.Client, device.ClientVersion),
		}); err != nil {
		defer func() {
			_ = cltx.Rollback(ctx)
		}()
		return nil, err
	}

	if err = s.IClientStore.SetClient(ctx, cltx, model.ClientSignUpRequest{
		Iaid:         Iaid,
		ClientId:     uuid.New().String(),
		ClientSecret: base64.StdEncoding.EncodeToString([]byte(uuid.New().String())),
	}); err != nil {
		defer func() {
			_ = utx.Rollback(ctx)
		}()
		return nil, err
	}

	if err = utx.Commit(ctx); err != nil {
		return nil, err
	}

	if err = cltx.Commit(ctx); err != nil {
		return nil, err
	}

	if err = s.RedisTc().SetToken(ctx, signature, accessToken, "access"); err != nil {
		return nil, err
	}

	if err = s.RedisTc().SetToken(ctx, signature, refreshToken, "refresh"); err != nil {
		return nil, err
	}

	return &model.SignUpDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Signature:    signature,
	}, nil
}

func (s *Service) approveSignupAsid(ctx *gin.Context, asid, login string) (bool, error) {
	loginAsid, err := s.IRedisDc.Get(ctx, asid)

	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return false, errors.New("no such asid")
		default:
			return false, err
		}
	}

	if loginAsid != login {
		return false, errors.New("asid login malformed")
	}

	return true, nil
}

func (s *Service) GenToken(Iaid, Eaid, oauthSecret, signature string, access ...bool) (string, error) {
	return jwt.NewAccessRefreshToken(
		model.AccessRefreshToken{
			Iaid:      Iaid,
			Eaid:      Eaid,
			Signature: signature,
		},
		oauthSecret,
		access...,
	)
}
