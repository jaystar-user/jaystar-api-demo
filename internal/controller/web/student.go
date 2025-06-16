package web

import (
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/controller/web/util"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/model/po"
	"jaystar/internal/utils/ctxUtil"
	"jaystar/internal/utils/errs"
	"jaystar/internal/utils/strUtil"
	"net/http"
	"strconv"
	"time"
)

func ProvideStudentController(studentSrv interfaces.IStudentSrv, reqParse util.IRequestParse, logger logger.ILogger) *StudentCtrl {
	return &StudentCtrl{
		studentSrv: studentSrv,
		reqParse:   reqParse,
		logger:     logger,
	}
}

type StudentCtrl struct {
	studentSrv interfaces.IStudentSrv
	reqParse   util.IRequestParse
	logger     logger.ILogger
}

func (ctrl *StudentCtrl) GetStudents(ctx *gin.Context) {
	req := &dto.StudentIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	userStore := ctxUtil.GetUserSessionFromCtx(ctx)

	boStudentCond := &bo.StudentCond{}
	boStudentCond.UserId = userStore.UserId

	boStudents, _, err := ctrl.studentSrv.GetStudents(ctx, boStudentCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	studentsVO := make([]dto.StudentVO, 0, len(boStudents))
	for _, boStudent := range boStudents {
		studentsVO = append(studentsVO, dto.StudentVO{
			StudentId:        strconv.FormatInt(boStudent.StudentId, 10),
			StudentRefId:     boStudent.StudentRefId,
			StudentName:      strUtil.GetStudentNameByStudentName(boStudent.StudentName),
			ParentName:       boStudent.ParentName,
			ParentPhone:      boStudent.ParentPhone,
			Balance:          boStudent.Balance,
			Mode:             boStudent.Mode.ToKey(),
			IsSettleNormally: boStudent.IsSettleNormally,
		})
	}

	SetStandardResponse(ctx, http.StatusOK, studentsVO)
}

func (ctrl *StudentCtrl) AdminGetStudents(ctx *gin.Context) {
	req := &dto.AdminGetStudentsIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boStudentCond := &bo.StudentCond{}
	if req.StudentName != nil {
		boStudentCond.StudentName = *req.StudentName
	}
	if req.ParentPhone != nil {
		boStudentCond.ParentPhone = *req.ParentPhone
	}
	if req.StudentRefId != nil {
		boStudentCond.StudentRefId = *req.StudentRefId
	}
	if req.Mode != nil {
		mode := kintone.ModeToEnum(*req.Mode)
		boStudentCond.Mode = &mode
	}
	boStudentCond.IsDeleted = req.IsDeleted
	boStudentCond.IsSettleNormally = req.IsSettleNormally
	if req.PagerIO != nil {
		boStudentCond.Pager = &po.Pager{}
		boStudentCond.Pager.Index = req.PagerIO.Index
		boStudentCond.Pager.Size = req.PagerIO.Size
	}

	students, pagerResult, err := ctrl.studentSrv.GetStudents(ctx, boStudentCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	studentsVO := make([]dto.AdminStudentVO, 0, len(students))
	for _, boStudent := range students {
		adminStudentVO := dto.AdminStudentVO{}
		adminStudentVO.StudentId = strconv.FormatInt(boStudent.StudentId, 10)
		adminStudentVO.StudentRefId = boStudent.StudentRefId
		adminStudentVO.StudentName = strUtil.GetStudentNameByStudentName(boStudent.StudentName)
		adminStudentVO.ParentName = boStudent.ParentName
		adminStudentVO.ParentPhone = boStudent.ParentPhone
		adminStudentVO.Balance = boStudent.Balance
		adminStudentVO.Mode = boStudent.Mode.ToKey()
		adminStudentVO.IsSettleNormally = boStudent.IsSettleNormally
		adminStudentVO.IsDeleted = boStudent.IsDeleted
		adminStudentVO.CreatedAt = boStudent.CreatedAt.Format(time.RFC3339)
		adminStudentVO.UpdatedAt = boStudent.UpdatedAt.Format(time.RFC3339)
		if boStudent.DeletedAt != nil {
			adminStudentVO.DeletedAt = boStudent.DeletedAt.Format(time.RFC3339)
		}

		studentsVO = append(studentsVO, adminStudentVO)
	}

	listVO := dto.ListVO{
		List: studentsVO,
		Pager: dto.PagerVO{
			Index: pagerResult.Index,
			Size:  pagerResult.Size,
			Pages: pagerResult.Pages,
			Total: pagerResult.Total,
		},
	}

	SetStandardResponse(ctx, http.StatusOK, listVO)
}

func (ctrl *StudentCtrl) KintoneStudentWebhook(ctx *gin.Context) {
	req := dto.KintoneWebhookStudentIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		ctrl.logger.Error(ctx, "studentCtrl KintoneStudentWebhook Bind", err)
		return
	}

	switch req.Type {
	case kintone.AddRecordType:
		ctrl.addStudent(ctx, req)
	case kintone.UpdateRecordType, kintone.UpdateStatusType:
		ctrl.updateStudent(ctx, req)
	case kintone.DeleteRecordType:
		ctrl.deleteStudent(ctx, req)
	}
}

func (ctrl *StudentCtrl) addStudent(ctx *gin.Context, req dto.KintoneWebhookStudentIO) {
	boStudent, err := ctrl.checkBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "studentCtrl addStudent checkBasicRequestData", err, zap.String("record_id", req.Record.Id.Value))
		return
	}

	err = ctrl.studentSrv.UserRegisterAndCreateStudent(ctx, boStudent)
	if err != nil {
		ctrl.logger.Error(ctx, "studentCtrl addStudent UserRegisterAndCreateStudent", err, zap.Int("record_ref_id", boStudent.StudentRefId))
	}
}

func (ctrl *StudentCtrl) updateStudent(ctx *gin.Context, req dto.KintoneWebhookStudentIO) {
	boStudent, err := ctrl.checkBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "studentCtrl updateStudent checkBasicRequestData", err, zap.String("record_id", req.Record.Id.Value))
		return
	}

	if err = ctrl.studentSrv.UpdateStudent(ctx, boStudent); err != nil {
		ctrl.logger.Error(ctx, "studentCtrl updateStudent UpdateStudentProcess", err, zap.Int("record_ref_id", boStudent.StudentRefId))
	}
}

func (ctrl *StudentCtrl) deleteStudent(ctx *gin.Context, req dto.KintoneWebhookStudentIO) {
	studentRefId, err := strconv.Atoi(req.RecordId)
	if err != nil {
		ctrl.logger.Error(ctx, "studentCtrl deleteStudent Atoi", err, zap.String("record_id", req.RecordId))
		return
	}

	if err = ctrl.studentSrv.DeleteStudent(ctx, studentRefId); err != nil {
		ctrl.logger.Error(ctx, "studentCtrl deleteStudent DeleteStudent", err, zap.Int("record_ref_id", studentRefId))
		return
	}
}

func (ctrl *StudentCtrl) checkBasicRequestData(req dto.KintoneWebhookStudentIO) (*bo.Student, error) {
	if req.Record.StudentName.Value == "" {
		return nil, xerrors.Errorf("studentCtrl checkBasicRequestData failed: %w", errs.StudentErr.StudentNameInvalidErr)
	}

	if req.Record.ParentPhone.Value == "" {
		return nil, xerrors.Errorf("studentCtrl checkBasicRequestData failed: %w", errs.StudentErr.ParentPhoneInvalidErr)
	}

	if req.Record.Id.Value == "" {
		return nil, xerrors.Errorf("studentCtrl checkBasicRequestData failed: %w", errs.StudentErr.StudentIdInvalidErr)
	}

	boStudent, err := req.Record.ToStudent()
	if err != nil {
		return nil, xerrors.Errorf("studentCtrl checkBasicRequestData req.Record.ToStudent: %w", err)
	}

	return boStudent, nil
}
