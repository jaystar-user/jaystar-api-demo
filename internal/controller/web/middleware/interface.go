package middleware

import "github.com/gin-gonic/gin"

type IMiddleware interface {
	Handle(ctx *gin.Context)
}
