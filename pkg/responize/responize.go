package responize

import (
	"github.com/gin-gonic/gin"
	"github.com/mxmrykov/aster-oauth-service/internal/model"
)

func R(ctx *gin.Context, data interface{}, code int, message string, error bool, refreshedToken ...string) {
	refreshedToken_ := new(string)

	if refreshedToken != nil {
		refreshedToken_ = &refreshedToken[0]
	}

	ctx.JSON(code, model.Response{
		Payload:        data,
		Status:         code,
		RefreshedToken: refreshedToken_,
		Message:        message,
		Error:          error,
	})
	ctx.Abort()
}
