package cronjob

import "jaystar/internal/constant/context"

func SetActionLogs(ctx *Context, key string, value any) {
	v, found := ctx.Get(context.ActionLogs)
	if !found {
		return
	}

	if actionLogs, ok := v.(map[string]any); ok {
		actionLogs[key] = value
	}
}
