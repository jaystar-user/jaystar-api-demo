package auth

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"jaystar/internal/constant/user"
	"time"
)

type UserSession struct {
	UserId int64
	Level  user.UserLevel // unused
}

func RegSessionValueTypes() {
	gob.Register(&UserSession{})
	gob.Register(time.Time{})
}

func GetUserSession(ctx *gin.Context, key string) *UserSession {
	session := sessions.Default(ctx)
	userStore := session.Get(key)
	if userStore == nil {
		return nil
	}

	if v, ok := userStore.(*UserSession); ok {
		return v
	}

	return nil
}

func RenewSession(ctx *gin.Context, env string) error {
	session := sessions.Default(ctx)

	// only using https cookie in production
	secure := false
	if env == "prod" {
		secure = true
	}

	session.Options(sessions.Options{
		Path:   "/",
		Domain: ctx.Request.URL.Host,
		MaxAge: 60 * 60,
		Secure: secure,
	})

	return session.Save()
}
