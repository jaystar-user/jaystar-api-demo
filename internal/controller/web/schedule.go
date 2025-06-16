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
	"jaystar/internal/service"
	"jaystar/internal/utils/ctxUtil"
	"jaystar/internal/utils/errs"
	"jaystar/internal/utils/timeUtil"
	"net/http"
	"strconv"
	"time"
)

func ProvideScheduleController(scheduleSrv interfaces.IScheduleSrv, reqParse util.IRequestParse, logger logger.ILogger) *ScheduleCtrl {
	return &ScheduleCtrl{
		scheduleSrv: scheduleSrv,
		reqParse:    reqParse,
		logger:      logger,
	}
}

type ScheduleCtrl struct {
	scheduleSrv interfaces.IScheduleSrv
	reqParse    util.IRequestParse
	logger      logger.ILogger
}

func (ctrl *ScheduleCtrl) GetSchedule(ctx *gin.Context) {
	req := &dto.ScheduleGetIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boScheduleCond := &bo.GetScheduleCond{
		ClassTimeStart: req.ClassTimeStart,
		ClassTimeEnd:   req.ClassTimeEnd,
		Pager: po.Pager{
			Index: req.Index,
			Size:  req.Size,
			Order: "class_time desc",
		},
	}

	userStore := ctxUtil.GetUserSessionFromCtx(ctx)

	boStudentCond := &bo.StudentCond{}
	boStudentCond.UserId = userStore.UserId
	if req.StudentName != nil {
		boStudentCond.StudentName = *req.StudentName

	}
	if req.ParentPhone != nil {
		boStudentCond.ParentPhone = *req.ParentPhone
	}

	boSchedules, pagerResult, err := ctrl.scheduleSrv.GetSchedules(ctx, boScheduleCond, boStudentCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	scheduleVO := make([]dto.ScheduleVO, 0, len(boSchedules))
	for _, schedule := range boSchedules {
		scheduleVO = append(scheduleVO, dto.ScheduleVO{
			Id:          strconv.FormatInt(schedule.ScheduleId, 10),
			TeacherName: schedule.TeacherName,
			StudentName: schedule.StudentName,
			ClassLevel:  schedule.ClassLevel.ToValue(),
			ClassType:   schedule.ClassType.ToValue(),
			ClassTime:   schedule.ClassTime.In(timeUtil.GetLocalLocation()).Format(time.DateTime),
		})
	}
	listVO := dto.ListVO{
		List: scheduleVO,
		Pager: dto.PagerVO{
			Index: pagerResult.Index,
			Size:  pagerResult.Size,
			Pages: pagerResult.Pages,
			Total: pagerResult.Total,
		},
	}

	SetStandardResponse(ctx, http.StatusOK, listVO)
}

func (ctrl *ScheduleCtrl) AdminGetSchedules(ctx *gin.Context) {
	req := &dto.AdminGetSchedulesIO{}
	if err := ctrl.reqParse.Bind(ctx, &req); err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, errs.CommonErr.RequestParamError)
		return
	}

	boScheduleCond := &bo.GetScheduleCond{
		Pager: po.Pager{
			Index: req.Index,
			Size:  req.Size,
			Order: "class_time desc, schedule_id desc",
		},
	}
	if req.ClassTimeStart != nil {
		boScheduleCond.ClassTimeStart = *req.ClassTimeStart
	}
	if req.ClassTimeEnd != nil {
		boScheduleCond.ClassTimeEnd = *req.ClassTimeEnd
	}
	boScheduleCond.IsDeleted = req.IsDeleted

	studentCond := &bo.StudentCond{}
	if req.StudentName != nil {
		studentCond.StudentName = *req.StudentName
	}
	if req.ParentPhone != nil {
		studentCond.ParentPhone = *req.ParentPhone
	}

	boSchedules, pagerResult, err := ctrl.scheduleSrv.GetSchedules(ctx, boScheduleCond, studentCond)
	if err != nil {
		SetStandardResponse(ctx, http.StatusBadRequest, err)
		return
	}

	scheduleVO := make([]dto.AdminGetScheduleVO, 0, len(boSchedules))
	for _, schedule := range boSchedules {
		adminGetScheduleVO := dto.AdminGetScheduleVO{}
		adminGetScheduleVO.Id = strconv.FormatInt(schedule.ScheduleId, 10)
		adminGetScheduleVO.TeacherName = schedule.TeacherName
		adminGetScheduleVO.StudentName = schedule.StudentName
		adminGetScheduleVO.ParentPhone = schedule.ParentPhone
		adminGetScheduleVO.ClassLevel = schedule.ClassLevel.ToValue()
		adminGetScheduleVO.ClassType = schedule.ClassType.ToValue()
		adminGetScheduleVO.ClassTime = schedule.ClassTime.In(timeUtil.GetLocalLocation()).Format(time.DateTime)
		adminGetScheduleVO.IsDeleted = schedule.IsDeleted
		adminGetScheduleVO.CreatedAt = schedule.CreatedAt.Format(time.RFC3339)
		adminGetScheduleVO.UpdatedAt = schedule.UpdatedAt.Format(time.RFC3339)
		if schedule.DeletedAt != nil {
			adminGetScheduleVO.DeletedAt = schedule.DeletedAt.Format(time.RFC3339)
		}

		scheduleVO = append(scheduleVO, adminGetScheduleVO)
	}

	listVo := dto.ListVO{
		List: scheduleVO,
		Pager: dto.PagerVO{
			Index: pagerResult.Index,
			Size:  pagerResult.Size,
			Pages: pagerResult.Pages,
			Total: pagerResult.Total,
		},
	}

	SetStandardResponse(ctx, http.StatusOK, listVo)
}

func (ctrl *ScheduleCtrl) KintoneScheduleWebhook(ctx *gin.Context) {
	req := dto.KintoneWebhookScheduleIO{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctrl.logger.Error(ctx, "reduceRecordCtrl KintoneScheduleWebhook Bind", err)
		return
	}

	switch req.Type {
	case kintone.AddRecordType:
		ctrl.addSchedule(ctx, req)
	case kintone.UpdateRecordType, kintone.UpdateStatusType:
		ctrl.updateSchedule(ctx, req)
	case kintone.DeleteRecordType:
		ctrl.deleteSchedule(ctx, req)
	}
}

func (ctrl *ScheduleCtrl) addSchedule(ctx *gin.Context, req dto.KintoneWebhookScheduleIO) {
	boKintoneSchedules, err := ctrl.checkBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "scheduleCtrl addSchedule checkBasicRequestData", err, zap.String("record_id", req.Record.Id.Value))
		return
	}

	failedRecords := make([]service.FailedRecord, 0, len(boKintoneSchedules))

	for _, boKintoneSchedule := range boKintoneSchedules {
		studentCond := &bo.StudentCond{
			StudentName: boKintoneSchedule.StudentName,
			ParentPhone: boKintoneSchedule.ParentPhone,
		}

		if err = ctrl.scheduleSrv.AddSchedule(ctx, boKintoneSchedule, studentCond); err != nil {
			ctrl.logger.Error(ctx, "scheduleCtrl addSchedule AddSchedule", err, zap.String("student_name", boKintoneSchedule.StudentName))
			failedRecords = append(failedRecords, service.FailedRecord{ScheduleRefId: boKintoneSchedule.ScheduleRefId, Name: boKintoneSchedule.StudentName})
		}
	}
	if len(failedRecords) > 0 {
		ctrl.logger.Warn(ctx, "added schedule failed records", zap.ObjectValues("records", failedRecords))
	}
}

func (ctrl *ScheduleCtrl) updateSchedule(ctx *gin.Context, req dto.KintoneWebhookScheduleIO) {
	boKintoneSchedules, err := ctrl.checkBasicRequestData(req)
	if err != nil {
		ctrl.logger.Error(ctx, "scheduleCtrl updateSchedule checkBasicRequestData", err, zap.String("record_id", req.Record.Id.Value))
		return
	}
	// 這邊不用檢查 ToId error 因為在 checkBasicRequestData 中的 ToSchedules 已經有檢查過
	scheduleRefId, _ := req.Record.Id.ToId()

	// 1. 找出 db 所有對應這筆 schedule id 的資料
	schedulesInDb, err := ctrl.scheduleSrv.GetSchedulesByRefId(ctx, scheduleRefId)
	if err != nil {
		ctrl.logger.Error(ctx, "scheduleCtrl updateSchedule GetSchedulesByRefId", err, zap.Int("schedule_ref_id", scheduleRefId))
		return
	}

	failedRecords := make([]service.FailedRecord, 0, len(boKintoneSchedules)+len(schedulesInDb))

	schedulesInDbMap := map[bo.StudentCond]*bo.Schedule{}
	schedulesInDbIdMap := map[int]*bo.Schedule{}
	currentRecordStudentNames := map[bo.StudentCond]struct{}{}
	for _, scheduleInDb := range schedulesInDb {
		schedulesInDbMap[bo.StudentCond{StudentName: scheduleInDb.StudentName, ParentPhone: scheduleInDb.ParentPhone}] = scheduleInDb
		if scheduleInDb.RecordRefId != nil && *scheduleInDb.RecordRefId != 0 {
			schedulesInDbIdMap[*scheduleInDb.RecordRefId] = scheduleInDb
		}
	}
	// 2. 新增 (kintone 有，db沒有)
	// 3. 更新 (kintone 有，db有)
	for _, kintoneSchedule := range boKintoneSchedules {
		// 新增
		if (kintoneSchedule.RecordRefId == nil || schedulesInDbIdMap[*kintoneSchedule.RecordRefId] == nil) && schedulesInDbMap[bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone}] == nil {
			err := ctrl.scheduleSrv.AddSchedule(ctx, kintoneSchedule, &bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone})
			if err != nil {
				ctrl.logger.Error(ctx, "scheduleCtrl updateSchedule AddSchedule", err, zap.Int("schedule_ref_id", kintoneSchedule.ScheduleRefId), zap.String("student_name", kintoneSchedule.StudentName))
				failedRecords = append(failedRecords, service.FailedRecord{ScheduleRefId: kintoneSchedule.ScheduleRefId, Name: kintoneSchedule.StudentName})
			}
		} else if kintoneSchedule.RecordRefId != nil && schedulesInDbIdMap[*kintoneSchedule.RecordRefId] != nil {
			// 更新 by record_ref_id
			if err := ctrl.scheduleSrv.UpdateScheduleById(
				ctx,
				schedulesInDbIdMap[*kintoneSchedule.RecordRefId].ScheduleId,
				&bo.StudentCond{StudentId: schedulesInDbIdMap[*kintoneSchedule.RecordRefId].StudentId},
				&bo.UpdateScheduleData{
					RecordRefId: kintoneSchedule.RecordRefId,
					ClassLevel:  kintoneSchedule.ClassLevel,
					ClassType:   kintoneSchedule.ClassType,
					ClassTime:   kintoneSchedule.ClassTime,
					TeacherName: kintoneSchedule.TeacherName,
				}); err != nil {
				ctrl.logger.Error(ctx, "scheduleCtrl updateSchedule by recordRefId by record_ref_id scheduleSrv.UpdateScheduleById", err, zap.Int("schedule_ref_id", kintoneSchedule.ScheduleRefId), zap.String("student_name", kintoneSchedule.StudentName))
				failedRecords = append(failedRecords, service.FailedRecord{ScheduleId: schedulesInDbMap[bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone}].ScheduleId, ScheduleRefId: kintoneSchedule.ScheduleRefId, Name: kintoneSchedule.StudentName})
			}
		} else if schedulesInDbMap[bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone}] != nil {
			// 更新 student_name
			if err := ctrl.scheduleSrv.UpdateScheduleById(
				ctx,
				schedulesInDbMap[bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone}].ScheduleId,
				&bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone},
				&bo.UpdateScheduleData{
					RecordRefId: kintoneSchedule.RecordRefId,
					ClassLevel:  kintoneSchedule.ClassLevel,
					ClassType:   kintoneSchedule.ClassType,
					ClassTime:   kintoneSchedule.ClassTime,
					TeacherName: kintoneSchedule.TeacherName,
				}); err != nil {
				ctrl.logger.Error(ctx, "scheduleCtrl updateSchedule by studentName and parentPhone UpdateScheduleById", err, zap.Int("schedule_ref_id", kintoneSchedule.ScheduleRefId), zap.String("student_name", kintoneSchedule.StudentName))
				failedRecords = append(failedRecords, service.FailedRecord{ScheduleId: schedulesInDbMap[bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone}].ScheduleId, ScheduleRefId: kintoneSchedule.ScheduleRefId, Name: kintoneSchedule.StudentName})
			}
		}
		// 記錄現有的 record_id
		currentRecordStudentNames[bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone}] = struct{}{}
	}
	// 4. 刪除 (kintone 沒有，db有)
	for _, scheduleInDb := range schedulesInDb {
		if _, found := currentRecordStudentNames[bo.StudentCond{StudentName: scheduleInDb.StudentName, ParentPhone: scheduleInDb.ParentPhone}]; !found {
			err := ctrl.scheduleSrv.DeleteSchedule(ctx, &bo.UpdateScheduleCond{ScheduleId: scheduleInDb.ScheduleId})
			if err != nil {
				ctrl.logger.Error(ctx, "scheduleCtrl updateSchedule DeleteSchedule", err, zap.Int("schedule_ref_id", scheduleInDb.ScheduleRefId))
				failedRecords = append(failedRecords, service.FailedRecord{ScheduleId: scheduleInDb.ScheduleId, ScheduleRefId: scheduleInDb.ScheduleRefId, Name: scheduleInDb.StudentName})
			}
		}
	}

	if len(failedRecords) > 0 {
		ctrl.logger.Warn(ctx, "updated schedule failed records", zap.ObjectValues("records", failedRecords))
	}
}

func (ctrl *ScheduleCtrl) deleteSchedule(ctx *gin.Context, req dto.KintoneWebhookScheduleIO) {
	// 刪除 db 此 scheduleId 對應的所有資料
	scheduleRefId, err := strconv.Atoi(req.RecordId)
	if err != nil {
		ctrl.logger.Error(ctx, "scheduleCtrl deleteSchedule Atoi", err, zap.String("record_id", req.RecordId))
		return
	}

	err = ctrl.scheduleSrv.DeleteSchedule(ctx, &bo.UpdateScheduleCond{ScheduleRefId: scheduleRefId})
	if err != nil {
		ctrl.logger.Error(ctx, "scheduleCtrl deleteSchedule DeleteSchedule", err, zap.Int("schedule_ref_id", scheduleRefId))
	}
}

func (ctrl *ScheduleCtrl) checkBasicRequestData(req dto.KintoneWebhookScheduleIO) ([]*bo.KintoneSchedule, error) {
	if req.RecordId == "" && req.Record.Id.Value == "" {
		return nil, xerrors.Errorf("reduceRecordCtrl checkBasicRequestData failed: %w", errs.ScheduleErr.InvalidScheduleRefIdError)
	}

	kintoneSchedules, err := req.Record.ToSchedules()
	if err != nil {
		return nil, xerrors.Errorf("scheduleCtrl checkBasicRequestData req.Record.ToSchedules: %w", err)
	}

	return kintoneSchedules, nil
}
