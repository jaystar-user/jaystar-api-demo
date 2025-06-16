package web

import (
	"github.com/gin-gonic/gin"
	"jaystar/internal/controller/web"
	"jaystar/internal/controller/web/middleware"
	"jaystar/internal/utils/auth"
)

type IWebApp interface {
	Init(g *gin.Engine)
}

func ProvideWebApp(
	respMw middleware.IResponseMiddleware,
	httpLogMw middleware.IHttpLogMiddleware,
	authMw middleware.IAuthMiddleware,
	recoverMw middleware.IRecoverMiddleware,
	internalAuthMw middleware.IInternalAuthMiddleware,
	ctrl *web.Controller,
) IWebApp {
	return &webApp{
		RespMw:         respMw,
		HttpLogMw:      httpLogMw,
		Ctrl:           ctrl,
		AuthMw:         authMw,
		RecoverMw:      recoverMw,
		InternalAuthMw: internalAuthMw,
	}
}

type webApp struct {
	Ctrl           *web.Controller
	RespMw         middleware.IResponseMiddleware
	HttpLogMw      middleware.IHttpLogMiddleware
	AuthMw         middleware.IAuthMiddleware
	RecoverMw      middleware.IRecoverMiddleware
	InternalAuthMw middleware.IInternalAuthMiddleware
}

func (app *webApp) Init(g *gin.Engine) {
	auth.RegSessionValueTypes()
	app.setInternalRoutes(g)
	app.setWebhookRoutes(g)
	app.setApiRoutes(g)
}
