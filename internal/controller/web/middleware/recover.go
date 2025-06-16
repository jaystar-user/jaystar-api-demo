package middleware

import (
	"fmt"
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/gin-gonic/gin"
	"golang.org/x/xerrors"
	"jaystar/internal/constant/context"
	"jaystar/internal/constant/log"
	"net/http"
	"runtime/debug"
)

type IRecoverMiddleware interface {
	IMiddleware
	WebhookHandle(ctx *gin.Context)
}

func ProvideRecoverMiddleware(logger logger.ILogger) IRecoverMiddleware {
	return &recoverMiddleware{
		logger: logger,
	}
}

type recoverMiddleware struct {
	logger logger.ILogger
}

func (m *recoverMiddleware) Handle(ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			// 印 log 在 http middleware
			ctx.Set(context.StackTrace, string(debug.Stack()))
			SetResp(ctx, http.StatusInternalServerError, xerrors.Errorf("RecoverMiddleware Handle error recover: %v", r))
		}
	}()

	ctx.Next()
}

func (m *recoverMiddleware) WebhookHandle(ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			ctx.Set(context.StackTrace, string(debug.Stack()))
			ctx.Set(context.ActionLogs, map[string]any{
				log.ErrorMessage: fmt.Sprintf("%v", r),
			})
		}
	}()

	ctx.Next()
}
