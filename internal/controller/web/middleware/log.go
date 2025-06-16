package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"jaystar/internal/constant/context"
	"jaystar/internal/constant/log"
	"jaystar/internal/constant/user"
	"jaystar/internal/utils"
	"net/http"
	"strings"
	"time"
)

type IHttpLogMiddleware interface {
	IMiddleware
}

func ProvideHttpLogMiddleware(logger logger.ILogger) IHttpLogMiddleware {
	return &httpLogMiddleware{
		logger: logger,
	}
}

type httpLogMiddleware struct {
	logger logger.ILogger
}

type bodyLogWriter struct {
	gin.ResponseWriter
	bodyBuffer *bytes.Buffer
}

func (blw bodyLogWriter) Write(b []byte) (int, error) {
	blw.bodyBuffer.Write(b)
	return blw.ResponseWriter.Write(b)
}

func (mw *httpLogMiddleware) Handle(ctx *gin.Context) {
	logMap := make(map[string]interface{})
	blw := &bodyLogWriter{bodyBuffer: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
	ctx.Writer = blw

	// before request
	mw.loggerStart(ctx, logMap)

	ctx.Next()

	// after request
	mw.loggerEnd(ctx, logMap, blw)
}

func (mw *httpLogMiddleware) loggerStart(ctx *gin.Context, m map[string]interface{}) {
	reqStartTime := time.Now()
	ctx.Set(context.RequestStartTime, reqStartTime)
	m[log.RequestTime] = reqStartTime.In(utils.GetLocation()).Format(time.RFC3339Nano)

	id := uuid.NewString()
	ctx.Set(logger.CtxActionIdKey, id)

	m[log.Method] = ctx.Request.Method

	m[log.Path] = ctx.Request.URL.Path

	m[log.ClientIP] = ctx.ClientIP()

	m[log.Query] = ctx.Request.URL.RawQuery

	m[log.UserAgent] = ctx.Request.UserAgent()

	m[log.ContentType] = ctx.Request.Header.Get("Content-Type")

	cookie, err := ctx.Request.Cookie(user.SessionID)
	if err == nil {
		m[log.SessionId] = cookie.String()
	}

	if ctx.Request.Method == http.MethodPost || ctx.Request.Method == http.MethodPut {
		body, _ := ctx.GetRawData()
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		ctx.Set(context.RequestBody, string(body))
	}
}

func (mw *httpLogMiddleware) loggerEnd(ctx *gin.Context, m map[string]interface{}, blw *bodyLogWriter) {
	m[log.Body] = ctx.GetString(context.RequestBody)

	reqStartTime := ctx.GetTime(context.RequestStartTime)
	usedTime := time.Since(reqStartTime)
	m[log.UsedTime] = usedTime.String()

	statusCode := ctx.Writer.Status()
	m[log.ResponseStatus] = statusCode

	if statusCode >= http.StatusBadRequest {
		responseBody := strings.Trim(blw.bodyBuffer.String(), "\n")
		if len(responseBody) > log.MaxPrintBodyLen {
			responseBody = responseBody[:log.MaxPrintBodyLen]
		}
		m[log.ResponseBody] = responseBody
	}

	m[log.StackTrace] = ctx.GetString(context.StackTrace)

	m[log.ServerTime] = time.Now().In(utils.GetLocation()).Format(time.RFC3339Nano)

	logMessages := mw.processLogMessage(ctx, m)
	mw.logger.Info(ctx, logMessages)
}

func (mw *httpLogMiddleware) processLogMessage(ctx *gin.Context, m map[string]interface{}) string {
	actionLogs := ctx.GetStringMap(context.ActionLogs)
	for k, v := range actionLogs {
		m[k] = v
	}

	logMessageByte, _ := json.Marshal(m)
	return string(logMessageByte)
}
