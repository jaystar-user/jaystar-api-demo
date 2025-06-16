package web

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"jaystar/internal/constant/user"
	"net/http"
)

var (
	authKey    = []byte{244, 195, 59, 48, 142, 194, 159, 37, 55, 24, 137, 251, 140, 115, 249, 129, 177, 108, 172, 184, 246, 216, 45, 224, 32, 77, 245, 48, 174, 210, 58, 47}
	encryptKey = []byte{218, 39, 222, 9, 180, 65, 105, 1, 82, 148, 220, 60, 154, 3, 153, 9, 11, 183, 232, 100, 139, 124, 63, 99, 2, 143, 88, 153, 28, 235, 112, 64}
)

func (app *webApp) setInternalRoutes(g *gin.Engine) {
	internalGroup := g.Group("/_internal")
	internalGroup.GET("/health", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"message": "ok"}) })
	internalGroup.Use(app.HttpLogMw.Handle)
	internalGroup.Use(app.RespMw.Handle)
	internalGroup.Use(app.RecoverMw.Handle)

	internalAuthGroup := internalGroup.Group("")
	internalAuthGroup.Use(app.InternalAuthMw.Handle)
	internalAuthGroup.POST("/generate/password", app.Ctrl.UserCtrl.AdminGetEncryptedPassword)

	internalAuthGroup.GET("/users", app.Ctrl.UserCtrl.AdminGetUsers)
	internalAuthGroup.PUT("/user/:user_id", app.Ctrl.UserCtrl.AdminUpdateUser)

	internalAuthGroup.GET("/students", app.Ctrl.StudentCtrl.AdminGetStudents)
	internalAuthGroup.GET("/deposit_records", app.Ctrl.DepositRecordCtrl.AdminGetDepositRecords)
	internalAuthGroup.GET("/reduce_records", app.Ctrl.ReduceRecordCtrl.AdminGetReduceRecords)
	internalAuthGroup.GET("/schedules", app.Ctrl.ScheduleCtrl.AdminGetSchedules)

	internalAuthGroup.GET("/semester_settle_records", app.Ctrl.SemesterSettleRecordCtrl.AdminGetSemesterSettleRecords)
	internalAuthGroup.POST("/semester_settle_record/settle", app.Ctrl.SemesterSettleRecordCtrl.AdminSemesterSettlePoints)

	internalAuthGroup.POST("/sync/student", app.Ctrl.SyncCtrl.AdminBatchSyncStudentsAndUsers)
	internalAuthGroup.POST("/sync/deposit_record", app.Ctrl.SyncCtrl.AdminBatchSyncDepositRecord)
	internalAuthGroup.POST("/sync/reduce_record", app.Ctrl.SyncCtrl.AdminBatchSyncReduceRecord)
	internalAuthGroup.POST("/sync/schedule", app.Ctrl.SyncCtrl.AdminBatchSyncSchedule)
	internalAuthGroup.POST("/sync/point_card", app.Ctrl.SyncCtrl.AdminBatchSyncPointCard)
	internalAuthGroup.POST("/sync/semester_settle_record", app.Ctrl.SyncCtrl.AdminBatchSyncSemesterSettleRecord)
	internalAuthGroup.POST("/sync/all/by_student", app.Ctrl.SyncCtrl.AdminSyncAllByStudent)
}

func (app *webApp) setWebhookRoutes(g *gin.Engine) {
	webHookGroup := g.Group("/_kintone/webhook")
	webHookGroup.Use(app.HttpLogMw.Handle)
	webHookGroup.Use(app.RecoverMw.WebhookHandle)
	webHookGroup.POST("/student", app.Ctrl.StudentCtrl.KintoneStudentWebhook)
	webHookGroup.POST("/deposit_record", app.Ctrl.DepositRecordCtrl.KintoneDepositRecordWebhook)
	webHookGroup.POST("/reduce_record", app.Ctrl.ReduceRecordCtrl.KintoneReduceRecordWebhook)
	webHookGroup.POST("/schedule", app.Ctrl.ScheduleCtrl.KintoneScheduleWebhook)
	webHookGroup.POST("/semester_settle_record", app.Ctrl.SemesterSettleRecordCtrl.KintoneSemesterSettleRecordWebhook)
	webHookGroup.POST("/point_card", app.Ctrl.PointCardCtrl.KintonePointCardWebhook)
}

func (app *webApp) setApiRoutes(g *gin.Engine) {
	store := cookie.NewStore(authKey, encryptKey)
	apiGroup := g.Group("/api")
	apiGroup.Use(app.HttpLogMw.Handle)
	apiGroup.Use(app.RespMw.Handle)
	apiGroup.Use(app.RecoverMw.Handle)
	apiGroup.Use(sessions.Sessions(user.SessionID, store))

	apiGroup.POST("/user/login", app.Ctrl.UserCtrl.UserLogin)
	apiGroup.POST("/user", app.Ctrl.UserCtrl.RegisterUser)

	authApiGroup := apiGroup.Group("")
	authApiGroup.Use(app.AuthMw.Handle)
	authApiGroup.GET("/user", app.Ctrl.UserCtrl.GetUser)
	authApiGroup.PUT("/user/:user_id/password", app.Ctrl.UserCtrl.UpdateUserPassword)
	authApiGroup.POST("/user/logout", app.Ctrl.UserCtrl.UserLogout)
	authApiGroup.GET("/student/list", app.Ctrl.StudentCtrl.GetStudents)
	authApiGroup.GET("/deposit_record/list", app.Ctrl.DepositRecordCtrl.GetDepositRecords)
	authApiGroup.GET("/reduce_record/list", app.Ctrl.ReduceRecordCtrl.GetReduceRecords)
	authApiGroup.GET("/schedule/list", app.Ctrl.ScheduleCtrl.GetSchedule)
}
