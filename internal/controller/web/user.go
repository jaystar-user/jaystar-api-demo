package web

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"jaystar/internal/config"
	"jaystar/internal/constant/user"
	"jaystar/internal/controller/web/util"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/model/po"
	"jaystar/internal/utils/auth"
	"jaystar/internal/utils/ctxUtil"
	"jaystar/internal/utils/errs"
	"jaystar/internal/utils/strUtil"
	"net/http"
	"strconv"
	"time"
)

func ProvideUserController(userSrv interfaces.IUserSrv, reqParseSrv util.IRequestParse, cfg config.IConfigEnv) *UserCtrl {
	return &UserCtrl{
		userSrv:     userSrv,
		reqParseSrv: reqParseSrv,
		cfg:         cfg,
	}
}

type UserCtrl struct {
	userSrv     interfaces.IUserSrv
	reqParseSrv util.IRequestParse
	cfg         config.IConfigEnv
}

func (ctrl *UserCtrl) AdminGetUsers(ctx *gin.Context) {
	req := dto.AdminGetUsersIO{}
	if err := ctrl.reqParseSrv.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boUserCond := &bo.UserCond{
		Pager: po.Pager{
			Index: req.Index,
			Size:  req.Size,
		},
	}

	if req.Accounts != nil {
		boUserCond.Accounts = strUtil.GetAccArrFromAccStr(*req.Accounts)
	}

	if req.Status != nil {
		if *req.Status {
			boUserCond.Status = user.Activate
		} else {
			boUserCond.Status = user.Deactivate
		}
	}

	if req.IsChangedPassword != nil {
		boUserCond.IsChangedPassword = req.IsChangedPassword
	}

	boUsers, pagerResult, err := ctrl.userSrv.GetUsers(ctx, boUserCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	usersVO := make([]dto.AdminUserVO, 0, len(boUsers))
	for _, boUser := range boUsers {
		usersVO = append(usersVO, dto.AdminUserVO{
			UserId:            strconv.FormatInt(boUser.UserId, 10),
			Account:           boUser.Account,
			Status:            boUser.Status.ToValue(),
			Level:             boUser.Level.ToValue(),
			IsChangedPassword: boUser.IsChangedPassword,
			CreatedAt:         boUser.CreatedAt.Format(time.DateTime),
			UpdatedAt:         boUser.UpdatedAt.Format(time.DateTime),
		})
	}

	listVO := dto.ListVO{
		List: usersVO,
		Pager: dto.PagerVO{
			Index: pagerResult.Index,
			Size:  pagerResult.Size,
			Pages: pagerResult.Pages,
			Total: pagerResult.Total,
		},
	}

	SetStandardResponse(ctx, http.StatusOK, listVO)
}

func (ctrl *UserCtrl) AdminUpdateUser(ctx *gin.Context) {
	userIdInt, valid := ctrl.validateUserId(ctx)
	if !valid {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	req := dto.UpdateUserIO{}
	if err := ctrl.reqParseSrv.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	// cond
	boUserCond := &bo.UserCond{UserId: userIdInt}

	// data
	boUpdateUserData := &bo.UpdateUserData{
		IsChangedPassword: req.IsChangedPassword,
		Password:          req.Password,
	}
	if req.Status != nil {
		if *req.Status {
			boUpdateUserData.Status = user.Activate
		} else {
			boUpdateUserData.Status = user.Deactivate
		}
	}

	if err := ctrl.userSrv.UpdateUser(ctx, boUserCond, boUpdateUserData); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, nil)
}

func (ctrl *UserCtrl) AdminGetEncryptedPassword(ctx *gin.Context) {
	req := dto.PasswordIO{}
	if err := ctrl.reqParseSrv.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	if req.Password != nil && *req.Password == "" {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	password, err := ctrl.userSrv.GetEncryptedPassword(ctx, req)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, password)
}

func (ctrl *UserCtrl) GetUser(ctx *gin.Context) {
	userStore := ctxUtil.GetUserSessionFromCtx(ctx)

	boUserCond := &bo.UserCond{}
	boUserCond.UserId = userStore.UserId
	boUser, err := ctrl.userSrv.GetUser(ctx, boUserCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	userVo := &dto.UserVO{
		UserId:            strconv.FormatInt(boUser.UserId, 10),
		Account:           boUser.Account,
		Level:             boUser.Level.ToValue(),
		IsChangedPassword: boUser.IsChangedPassword,
	}

	SetStandardResponse(ctx, http.StatusOK, userVo)
}

func (ctrl *UserCtrl) UserLogin(ctx *gin.Context) {
	req := dto.UserLoginIO{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boUserLoginCond := &bo.UserLoginCond{
		Account:  req.Account,
		Password: req.Password,
	}

	boUser, err := ctrl.userSrv.UserLogin(ctx, boUserLoginCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	// only using https cookie in production
	secure := false
	if ctrl.cfg.GetAppEnv() == "prod" {
		secure = true
	}

	session := sessions.Default(ctx)
	session.Options(sessions.Options{
		Path:   "/",
		Domain: ctx.Request.URL.Host,
		MaxAge: 60 * 60,
		Secure: secure,
	})

	userSession := &auth.UserSession{
		UserId: boUser.UserId,
		Level:  boUser.Level,
	}
	session.Set(user.SessionUserKey, userSession)

	if err := session.Save(); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, nil)
}

func (ctrl *UserCtrl) UserLogout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Clear()

	// only using https cookie in production
	secure := false
	if ctrl.cfg.GetAppEnv() == "prod" {
		secure = true
	}

	session.Options(sessions.Options{
		Path:   "/",
		Domain: ctx.Request.URL.Host,
		MaxAge: -1,
		Secure: secure,
	})
	if err := session.Save(); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}
	SetStandardResponse(ctx, http.StatusOK, nil)
}

// RegisterUser 根據傳入的學生姓名、家長電話創建使用者帳號
func (ctrl *UserCtrl) RegisterUser(ctx *gin.Context) {
	req := dto.UserRegIO{}
	if err := ctrl.reqParseSrv.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boStudent := &bo.StudentCond{
		StudentName: req.StudentName,
		ParentPhone: req.ParentPhone,
	}

	boUser := &bo.CreateUserData{
		Account:  req.Account,
		Password: req.Password,
	}

	if err := ctrl.userSrv.UserRegister(ctx, boStudent, boUser); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, nil)
}

// UpdateUserPassword 第一次修改密碼
func (ctrl *UserCtrl) UpdateUserPassword(ctx *gin.Context) {
	userIdInt, valid := ctrl.validateUserId(ctx)
	if !valid {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	req := dto.UpdateUserPasswordIO{}
	if err := ctrl.reqParseSrv.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	// cond
	boUserCond := &bo.UserCond{UserId: userIdInt}

	// data
	isChangedPassword := true
	boUpdateUserData := &bo.UpdateUserData{
		IsChangedPassword: &isChangedPassword,
	}

	if req.Password != "" {
		boUpdateUserData.Password = &req.Password
	}

	if err := ctrl.userSrv.UpdateUser(ctx, boUserCond, boUpdateUserData); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	SetStandardResponse(ctx, http.StatusOK, nil)
}

func (ctrl *UserCtrl) validateUserId(ctx *gin.Context) (int64, bool) {
	userId := ctx.Param("user_id")
	if userId == "" {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return 0, false
	}
	userIdInt, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return 0, false
	}

	return userIdInt, true
}
