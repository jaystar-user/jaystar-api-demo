package middleware

import (
	"github.com/gin-gonic/gin"
	"jaystar/internal/constant/auth"
	"jaystar/internal/utils/errs"
	"net/http"
)

type IInternalAuthMiddleware interface {
	IMiddleware
}

func ProvideInternalAuthMiddleware() IInternalAuthMiddleware {
	return &internalAuthMiddleware{}
}

type internalAuthMiddleware struct{}

func (m *internalAuthMiddleware) Handle(ctx *gin.Context) {
	authKey := ctx.Request.Header.Get("Authorization")
	if authKey != auth.InternalAuthKey {
		SetResp(ctx, http.StatusUnauthorized, errs.CommonErr.AuthDeniedError)
		ctx.Abort()
		return
	}

	ctx.Next()
}
