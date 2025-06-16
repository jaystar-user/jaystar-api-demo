package middleware

import (
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/gin-gonic/gin"
	"jaystar/internal/config"
	"jaystar/internal/constant/context"
	"jaystar/internal/constant/user"
	"jaystar/internal/utils/auth"
	"jaystar/internal/utils/errs"
	"net/http"
)

type IAuthMiddleware interface {
	IMiddleware
}

func ProvideAuthMiddleware(logger logger.ILogger, config config.IConfigEnv) IAuthMiddleware {
	return &authMiddleware{
		logger: logger,
		cfg:    config,
	}
}

type authMiddleware struct {
	logger logger.ILogger
	cfg    config.IConfigEnv
}

func (m *authMiddleware) Handle(ctx *gin.Context) {
	store := auth.GetUserSession(ctx, user.SessionUserKey)
	if store == nil {
		SetResp(ctx, http.StatusUnauthorized, errs.CommonErr.AuthFailedError)
		ctx.Abort()
		return
	}

	ctx.Set(context.UserSession, *store)

	if err := auth.RenewSession(ctx, m.cfg.GetAppEnv()); err != nil {
		m.logger.Error(ctx, "authMiddleware auth.RenewSession", err)
	}

	ctx.Next()
}
