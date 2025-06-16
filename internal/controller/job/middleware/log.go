package middleware

import (
	"encoding/json"
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/google/uuid"
	"jaystar/internal/constant/context"
	"jaystar/internal/constant/log"
	"jaystar/internal/cronjob"
	"jaystar/internal/utils"
	"time"
)

type IJobMiddleware interface {
	Handle(ctx *cronjob.Context)
}

func ProvideJobLogMiddleware(logger logger.ILogger) *JobLogMiddleware {
	return &JobLogMiddleware{
		logger: logger,
	}
}

type JobLogMiddleware struct {
	logger logger.ILogger
}

func (mw *JobLogMiddleware) Handle(ctx *cronjob.Context) {
	logMap := make(map[string]interface{})

	// before request
	mw.loggerStart(ctx, logMap)

	ctx.Next()

	// after request
	mw.loggerEnd(ctx, logMap)
}

func (mw *JobLogMiddleware) loggerStart(ctx *cronjob.Context, m map[string]interface{}) {
	now := time.Now()
	m[log.RequestTime] = now.In(utils.GetLocation()).Format(time.RFC3339Nano)
	m[log.Type] = "job"

	ctx.Set(logger.CtxActionIdKey, uuid.NewString())
	ctx.Set(context.ActionLogs, m)
	ctx.Set(context.RequestStartTime, now)
}

func (mw *JobLogMiddleware) loggerEnd(ctx *cronjob.Context, m map[string]interface{}) {
	reqStartTime, found := ctx.Get(context.RequestStartTime)
	if v, ok := reqStartTime.(time.Time); found && ok {
		usedTime := time.Since(v)
		m[log.UsedTime] = usedTime.String()
	}

	m[log.ServerTime] = time.Now().In(utils.GetLocation()).Format(time.RFC3339Nano)

	logMessages := mw.processLogMessage(ctx, m)
	mw.logger.Info(ctx, logMessages)
}

func (mw *JobLogMiddleware) processLogMessage(ctx *cronjob.Context, m map[string]interface{}) string {
	v, found := ctx.Get(context.ActionLogs)
	if actionLogs, ok := v.(map[string]any); found && ok {
		for k, v := range actionLogs {
			m[k] = v
		}
	}

	logMessageByte, _ := json.Marshal(m)
	return string(logMessageByte)
}
