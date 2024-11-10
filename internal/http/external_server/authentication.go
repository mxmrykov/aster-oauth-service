package external_server

import (
	"github.com/gin-gonic/gin"
	"github.com/mxmrykov/aster-oauth-service/internal/model"
	"github.com/mxmrykov/aster-oauth-service/pkg/responize"
	"net/http"
	"strconv"
)

func (s *Server) authHandshake(ctx *gin.Context) {

}

func (s *Server) getPhoneCode(ctx *gin.Context) {
	request := new(model.GetPhoneCodeRequest)

	if err := ctx.ShouldBindQuery(&request); err != nil {
		s.svc.Logger().Err(err).Msg("Failed to bind query params")
		responize.R(ctx, nil, http.StatusBadRequest, "Invalid request", true)
		return
	}

	e, err := s.svc.IfPhoneInUse(ctx, request.Phone)

	switch {
	case err != nil:
		s.svc.Logger().Err(err).Msg("Failed to check if phone is in use")
		responize.R(ctx, nil, http.StatusInternalServerError, "Failed to check if phone is in use", true)
		return
	case e:
		s.svc.Logger().Err(err).Msg("Phone already in use")
		responize.R(ctx, nil, http.StatusBadRequest, "Phone already in use", true)
		return
	}

	if err = s.svc.SetPhoneConfirmCode(ctx, request.Phone); err != nil {
		s.svc.Logger().Err(err).Msg("Failed to set phone confirm code")
		responize.R(ctx, nil, http.StatusInternalServerError, "Failed to set phone confirm code", true)
		return
	}

	responize.R(ctx, nil, http.StatusOK, "Code was sent", false)
}

func (s *Server) signupHandshake(ctx *gin.Context) {
	_, mwLogin, request := ctx.GetString("asid"), ctx.GetString("login"), new(model.SignupRequest)

	if err := ctx.ShouldBindJSON(&request); err != nil {
		s.svc.Logger().Err(err).Msg("Failed to bind JSON")
		responize.R(ctx, nil, http.StatusBadRequest, "Invalid request", true)
		return
	}

	if mwLogin != request.Login {
		s.svc.Logger().Error().Msg("Incorrect token login")
		responize.R(ctx, nil, http.StatusBadRequest, "Invalid request", true)
		return
	}

	e, err := s.svc.IfPhoneInUse(ctx, request.Phone)

	switch {
	case err != nil:
		s.svc.Logger().Err(err).Msg("Failed to check if phone is in use")
		responize.R(ctx, nil, http.StatusInternalServerError, "Failed to check if phone is in use", true)
		return
	case e:
		s.svc.Logger().Err(err).Msg("Phone already in use")
		responize.R(ctx, nil, http.StatusBadRequest, "Phone already in use", true)
		return
	}

	e, err = s.svc.IfLoginInUse(ctx, request.Login)

	switch {
	case err != nil:
		s.svc.Logger().Err(err).Msg("Failed to check if login is in use")
		responize.R(ctx, nil, http.StatusInternalServerError, "Failed to check if login is in use", true)
		return
	case e:
		s.svc.Logger().Err(err).Msg("Login already in use")
		responize.R(ctx, nil, http.StatusBadRequest, "Login already in use", true)
		return
	}
}

func (s *Server) confirmCode(ctx *gin.Context) {
	request := new(model.ConfirmPhoneCodeRequest)

	if err := ctx.ShouldBindQuery(&request); err != nil {
		s.svc.Logger().Err(err).Msg("Failed to bind query params")
		responize.R(ctx, nil, http.StatusBadRequest, "Invalid request", true)
		return
	}

	e, err := s.svc.IfPhoneInUse(ctx, request.Phone)

	switch {
	case err != nil:
		s.svc.Logger().Err(err).Msg("Failed to check if phone is in use")
		responize.R(ctx, nil, http.StatusInternalServerError, "Failed to check if phone is in use", true)
		return
	case e:
		s.svc.Logger().Err(err).Msg("Phone already in use")
		responize.R(ctx, nil, http.StatusBadRequest, "Phone already in use", true)
		return
	}

	code, err := s.svc.GetPhoneConfirmCode(ctx, request.Phone)

	switch {
	case err != nil:
		s.svc.Logger().Err(err).Msg("Failed to get phone confirm code")
		responize.R(ctx, nil, http.StatusInternalServerError, "Failed to get phone confirm code", true)
		return
	case strconv.Itoa(request.Code) != code:
		s.svc.Logger().Err(err).Msg("Incorrect confirm code")
		responize.R(ctx, nil, http.StatusBadRequest, "Incorrect confirm code", true)
		return
	}

	responize.R(ctx, nil, http.StatusOK, "Code confirmed", false)
}
