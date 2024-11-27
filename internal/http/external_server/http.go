package external_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/mxmrykov/aster-oauth-service/internal/model"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mxmrykov/aster-oauth-service/internal/cache"
	"github.com/mxmrykov/aster-oauth-service/internal/config"
	"github.com/mxmrykov/aster-oauth-service/internal/store/postgres"
	"github.com/mxmrykov/aster-oauth-service/internal/store/redis"
	"github.com/mxmrykov/aster-oauth-service/pkg/clients/vault"
	"github.com/rs/zerolog"
)

type IServer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	IVault() vault.IVault
	ICache() cache.ICache
	OAuth() *config.OAuth
	Logger() *zerolog.Logger
	ClientStore() postgres.IClientStore
	UserStore() postgres.IUserStore
	RedisDc() redis.IRedisDc
	RedisTc() redis.IRedisTc

	SetPhoneConfirmCode(ctx *gin.Context, phone string) error
	IfCodeSent(ctx *gin.Context, phone string) (bool, error)
	IfPhoneInUse(ctx *gin.Context, phone string) (bool, error)
	IfLoginInUse(ctx *gin.Context, login string) (bool, error)
	GetPhoneConfirmCode(ctx *gin.Context, phone string) (string, error)
	SetPhoneConfirmed(ctx *gin.Context, phone string) error
	ValidateUserSignup(ctx *gin.Context, r *model.SignupRequest) error
	SignupUser(ctx *gin.Context, r *model.SignupRequest) (*model.AuthDTO, error)

	GenToken(Iaid, Eaid, oauthSecret, signature string, access ...bool) (string, error)
	Exit(ctx *gin.Context, signature, iaid string, id int)
	ValidateClientAuth(ctx context.Context, r *model.AuthRequest, iaid string) error
	ResourceOwnerAuthorize(ctx *gin.Context, iaid string) (*model.AuthDTO, error)
}

type Server struct {
	svc    IServer
	router *gin.Engine
	http   http.Server
}

const (
	// authenticationGroupV1 - неавторизованные пользователи, работаем с клиентами
	authenticationGroupV1 = "oauth/api/v1/authentication"
	// authorizationGroupV1 - авторизованные пользователи, работаем с токенами
	authorizationGroupV1 = "oauth/api/v1/authorization"

	authorizationEndpoint              = "/handshake"
	registrationEndpoint               = "/handshake"
	registrationGetConfirmCodeEndpoint = "/confirm/code"

	exitSessionEndpoint = "/exit/session"
)

func NewServer(logger *zerolog.Logger, svc IServer) *Server {
	router := gin.New()

	router.Use(
		gin.Logger(),
		gin.CustomRecoveryWithWriter(nil, recoveryFunc(logger)),
	)

	s := &Server{
		svc:    svc,
		router: router,
		http: http.Server{
			Addr:    fmt.Sprintf(":%d", svc.OAuth().ExternalServer.Port),
			Handler: router,
		},
	}

	s.configureRouter()

	return s
}

func (s *Server) configureRouter() {
	s.router.Use(s.footPrintAuth)
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://aster.ru"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-type", "X-TempAuth-Token", "X-Access-Token", "X-Auth-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	authenticationGroup := s.router.Group(authenticationGroupV1)

	aauth := authenticationGroup.Group("/auth")
	aauth.Use(s.internalAuthMiddleWare)
	aauth.POST(authorizationEndpoint, s.authHandshake)

	asignup := authenticationGroup.Group("/signup")
	asignup.Use(s.authenticationMw)
	asignup.POST(registrationEndpoint, s.signupHandshake)
	asignup.GET(registrationGetConfirmCodeEndpoint, s.getPhoneCode)
	asignup.POST(registrationGetConfirmCodeEndpoint, s.confirmCode)

	authorizationGroup := s.router.Group(authorizationGroupV1)
	authorizationGroup.Use(s.authorizationMw)
	authorizationGroup.POST(exitSessionEndpoint, s.exitSession)

}

func recoveryFunc(logger *zerolog.Logger) gin.RecoveryFunc {
	return func(c *gin.Context, err any) {
		logger.Error().Err(fmt.Errorf("PANIC: %v", err)).Send()
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

func (s *Server) Start(_ context.Context) error {
	if err := s.http.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
