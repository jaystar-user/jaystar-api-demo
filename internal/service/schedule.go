package service

import (
	"context"
	"errors"
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/SeanZhenggg/go-utils/snowflake/autoId"
	"github.com/panjf2000/ants"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/constant/request"
	"jaystar/internal/database"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/model/po"
	"jaystar/internal/utils/errs"
	"jaystar/internal/utils/pool"
	"jaystar/internal/utils/strUtil"
	"reflect"
	"sync"
	"time"
)

type FailedRecord struct {
	ScheduleId    int64
	ScheduleRefId int
	Name          string
}

func (f *FailedRecord) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddInt64("schedule_id", f.ScheduleId)
	encoder.AddInt("schedule_ref_id", f.ScheduleRefId)
	encoder.AddString("name", f.Name)
	return nil
}

func ProvideScheduleService(
	studentCommonSrv interfaces.IStudentCommonSrv,
	scheduleRepo interfaces.IScheduleRepo,
	db database.IPostgresDB,
	logger logger.ILogger,
	scheduleCommonSrv interfaces.IScheduleCommonSrv,
) *ScheduleService {
	return &ScheduleService{
		studentCommonSrv:  studentCommonSrv,
		scheduleRepo:      scheduleRepo,
		db:                db,
		logger:            logger,
		scheduleCommonSrv: scheduleCommonSrv,
		executorPool:      pool.NewExecutorPool(100),
	}
}

type ScheduleService struct {
	studentCommonSrv  interfaces.IStudentCommonSrv
	scheduleRepo      interfaces.IScheduleRepo
	db                database.IPostgresDB
	logger            logger.ILogger
	scheduleCommonSrv interfaces.IScheduleCommonSrv
	executorPool      *ants.Pool `wire:"-"`
}

func (srv *ScheduleService) GetSchedules(ctx context.Context, cond *bo.GetScheduleCond, studentCond *bo.StudentCond) ([]*bo.Schedule, *po.PagerResult, error) {
	poGetScheduleCond := &po.GetScheduleCond{}

	if !cond.ClassTimeStart.IsZero() {
		poGetScheduleCond.ClassTimeStart = cond.ClassTimeStart
	}
	if !cond.ClassTimeEnd.IsZero() {
		poGetScheduleCond.ClassTimeEnd = cond.ClassTimeEnd.Add(24 * time.Hour).Add(-1 * time.Minute)
	}
	if cond.IsDeleted != nil {
		poGetScheduleCond.IsDeleted = cond.IsDeleted
	} else {
		isDeleted := false
		poGetScheduleCond.IsDeleted = &isDeleted
	}

	var (
		boStudents *bo.Student
		err        error
	)
	// 有帶 student 條件就以學生為基礎撈資料
	if !reflect.ValueOf(studentCond).Elem().IsZero() {
		boStudents, err = srv.studentCommonSrv.GetStudent(ctx, studentCond)
		if err != nil {
			if errors.Is(err, errs.DbErr.NoRow) {
				return nil, nil, xerrors.Errorf("scheduleService GetSchedules studentCommonSrv.GetStudents: %w", errs.StudentErr.StudentNotFoundErr)
			}
			return nil, nil, xerrors.Errorf("scheduleService GetSchedules studentCommonSrv.GetStudents: %w", err)
		}

		poGetScheduleCond.StudentId = boStudents.StudentId
	}

	poPager := &po.Pager{
		Index: cond.Index,
		Size:  cond.Size,
		Order: cond.Order,
	}

	db := srv.db.Session()
	poSchedules, err := srv.scheduleRepo.GetSchedules(ctx, db, poGetScheduleCond, poPager)
	if err != nil {
		return nil, nil, xerrors.Errorf("scheduleService GetSchedules scheduleRepo.GetSchedules: %w", err)
	}
	poPagerResult, err := srv.scheduleRepo.GetSchedulesPager(ctx, db, poGetScheduleCond, poPager)
	if err != nil {
		return nil, nil, xerrors.Errorf("scheduleService GetSchedules scheduleRepo.GetRecordsPager: %w", err)
	}

	boSchedules := make([]*bo.Schedule, 0, len(poSchedules))
	for _, record := range poSchedules {
		boSchedules = append(boSchedules, &bo.Schedule{
			ScheduleId:    record.ScheduleId,
			ScheduleRefId: record.ScheduleRefId,
			StudentName:   record.StudentName,
			ParentPhone:   record.ParentPhone,
			RecordRefId:   record.RecordRefId,
			ClassLevel:    kintone.ClassLevelToEnum(record.ClassLevel),
			ClassType:     kintone.ClassTypeToEnum(record.ClassType),
			ClassTime:     *record.ClassTime,
			TeacherName:   record.TeacherName,
			IsDeleted:     record.IsDeleted,
			CreatedAt:     record.CreatedAt,
			UpdatedAt:     record.UpdatedAt,
			DeletedAt:     record.DeletedAt,
		})
	}

	return boSchedules, poPagerResult, nil
}

func (srv *ScheduleService) GetSchedulesByRefId(ctx context.Context, scheduleRefId int) ([]*bo.Schedule, error) {
	db := srv.db.Session()
	poSchedules, err := srv.scheduleRepo.GetSchedulesByRefId(ctx, db, scheduleRefId)
	if err != nil && !errors.Is(err, errs.DbErr.NoRow) {
		return nil, xerrors.Errorf("scheduleService GetSchedulesByRefId: %w", err)
	}

	if len(poSchedules) == 0 {
		return nil, nil
	}

	boSchedules := make([]*bo.Schedule, 0, len(poSchedules))
	for _, record := range poSchedules {
		boSchedules = append(boSchedules, &bo.Schedule{
			ScheduleId:    record.ScheduleId,
			ScheduleRefId: record.ScheduleRefId,
			RecordRefId:   record.RecordRefId,
			StudentId:     record.StudentId,
			StudentName:   record.StudentName,
			ParentPhone:   record.ParentPhone,
			ClassLevel:    kintone.ClassLevelToEnum(record.ClassLevel),
			ClassType:     kintone.ClassTypeToEnum(record.ClassType),
			ClassTime:     *record.ClassTime,
			TeacherName:   record.TeacherName,
			CreatedAt:     record.CreatedAt,
			UpdatedAt:     record.UpdatedAt,
		})
	}

	return boSchedules, nil
}

func (srv *ScheduleService) AddSchedule(ctx context.Context, data *bo.KintoneSchedule, studentCond *bo.StudentCond) error {
	if len(studentCond.StudentName) == 0 || len(studentCond.ParentPhone) == 0 {
		return errs.ScheduleErr.InvalidStudentNameErr
	}

	boStudent, err := srv.studentCommonSrv.GetStudent(ctx, studentCond)
	if err != nil {
		if errors.Is(err, errs.DbErr.NoRow) {
			return xerrors.Errorf("scheduleService AddSchedule studentCommonSrv.GetStudent: %w", errs.StudentErr.StudentNotFoundErr)
		}

		return xerrors.Errorf("scheduleService AddSchedule studentCommonSrv.GetStudent: %w", err)
	}

	scheduleId, err := autoId.DefaultSnowFlake.GenNextId()
	if err != nil {
		return xerrors.Errorf("scheduleService AddSchedule autoId.DefaultSnowFlake.GenNextId: %w", err)
	}

	poSchedule := &po.Schedule{
		ScheduleId:    scheduleId,
		ScheduleRefId: data.ScheduleRefId,
		RecordRefId:   data.RecordRefId,
		StudentId:     boStudent.StudentId,
		ClassLevel:    data.ClassLevel.ToKey(),
		ClassType:     data.ClassType.ToKey(),
		ClassTime:     &data.ClassTime,
		TeacherName:   data.TeacherName,
	}

	db := srv.db.Session()
	if err := srv.scheduleRepo.AddSchedule(ctx, db, poSchedule); err != nil {
		return xerrors.Errorf("scheduleService scheduleRepo scheduleRepo.AddSchedule: %w", err)
	}

	return nil
}

func (srv *ScheduleService) UpdateScheduleById(ctx context.Context, scheduleId int64, studentCond *bo.StudentCond, data *bo.UpdateScheduleData) error {
	if len(studentCond.StudentName) == 0 || len(studentCond.ParentPhone) == 0 {
		return errs.ScheduleErr.InvalidStudentNameErr
	}

	boStudent, err := srv.studentCommonSrv.GetStudent(ctx, studentCond)
	if err != nil {
		if errors.Is(err, errs.DbErr.NoRow) {
			return xerrors.Errorf("scheduleService UpdateSchedule studentCommonSrv.GetStudent: %w", errs.StudentErr.StudentNotFoundErr)
		}

		return xerrors.Errorf("scheduleService UpdateSchedule studentCommonSrv.GetStudent: %w", err)
	}

	db := srv.db.Session()
	// conditions
	poUpdateScheduleCond := &po.UpdateScheduleCond{}
	if scheduleId != 0 {
		poUpdateScheduleCond.ScheduleId = scheduleId
	}
	// data
	poUpdateScheduleData := &po.UpdateScheduleData{
		StudentId:   boStudent.StudentId,
		RecordRefId: data.RecordRefId,
		ClassType:   data.ClassType.ToKey(),
		ClassLevel:  data.ClassLevel.ToKey(),
		ClassTime:   &data.ClassTime,
		TeacherName: data.TeacherName,
	}

	if err := srv.scheduleRepo.UpdateSchedule(ctx, db, poUpdateScheduleCond, poUpdateScheduleData); err != nil {
		return xerrors.Errorf("scheduleService UpdateSchedule scheduleRepo.UpdateRecord: %w", err)
	}

	return nil
}

func (srv *ScheduleService) DeleteSchedule(ctx context.Context, cond *bo.UpdateScheduleCond) error {
	if cond.ScheduleId == 0 && cond.ScheduleRefId == 0 && cond.RecordRefId == 0 {
		return errs.CommonErr.CondError
	}

	deleted := true
	// conditions
	poUpdateScheduleCond := &po.UpdateScheduleCond{}
	if cond.ScheduleId != 0 {
		poUpdateScheduleCond.ScheduleId = cond.ScheduleId
	}
	if cond.ScheduleRefId != 0 {
		poUpdateScheduleCond.ScheduleRefId = cond.ScheduleRefId
	}
	if cond.RecordRefId != 0 {
		poUpdateScheduleCond.RecordRefId = cond.RecordRefId
	}
	// data
	now := time.Now()
	poUpdateScheduleData := &po.UpdateScheduleData{
		IsDeleted: &deleted,
		DeletedAt: &now,
	}

	db := srv.db.Session()
	if err := srv.scheduleRepo.UpdateSchedule(ctx, db, poUpdateScheduleCond, poUpdateScheduleData); err != nil {
		return xerrors.Errorf("scheduleService DeleteSchedule scheduleRepo.UpdateSchedule: %w", err)
	}

	return nil
}

func (srv *ScheduleService) BatchSyncSchedule(ctx context.Context, cond *bo.SyncScheduleCond, wg ...*sync.WaitGroup) error {
	var (
		total            int
		err              error
		kintoneSchedules []dto.ScheduleRecord
	)

	boScheduleReq := &dto.ScheduleReq{
		Limit:  request.GetRecordsBatchLimit,
		Offset: request.GetRecordsBatchOffset,
	}
	if cond != nil {
		if cond.StudentName != nil && cond.ParentPhone != nil {
			boScheduleReq.StudentName = strUtil.GetFullStudentName(*cond.StudentName, *cond.ParentPhone)
		}
		if cond.ClassTimeStart != nil {
			boScheduleReq.ClassTimeStart = *cond.ClassTimeStart
		}
		if cond.ClassTimeEnd != nil {
			boScheduleReq.ClassTimeEnd = *cond.ClassTimeEnd
		}
		if cond.Limit != nil {
			boScheduleReq.Limit = *cond.Limit
		}
		if cond.Offset != nil {
			boScheduleReq.Offset = *cond.Offset
		}
	}

	kintoneSchedules, total, err = srv.scheduleCommonSrv.GetKintoneSchedules(ctx, boScheduleReq)
	if err != nil {
		return xerrors.Errorf("scheduleService BatchSyncSchedule getKintoneSchedule: %w", err)
	}

	allRecords := make([]dto.ScheduleRecord, 0, total)
	allRecords = append(allRecords, kintoneSchedules...)

	offset, limit := boScheduleReq.Offset, boScheduleReq.Limit
	for ((offset * limit) + limit) < total {
		offset += 1
		boScheduleReq.Offset = offset * limit
		kintoneSchedules, _, err = srv.scheduleCommonSrv.GetKintoneSchedules(ctx, boScheduleReq)
		if err != nil {
			return xerrors.Errorf("scheduleService BatchSyncSchedule getKintoneSchedule: %w", err)
		}
		allRecords = append(allRecords, kintoneSchedules...)
	}

	var wait *sync.WaitGroup
	if len(wg) > 0 {
		wait = wg[0]
		wait.Add(1)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				srv.logger.Error(ctx, "scheduleService BatchSyncSchedule panic", nil, zap.Any(logger.PanicMessage, r))
			}
			if wait != nil {
				wait.Done()
			}
		}()

		srv.syncSchedules(ctx, allRecords, cond)
	}()
	return nil
}

func (srv *ScheduleService) syncSchedules(ctx context.Context, allRecords []dto.ScheduleRecord, cond *bo.SyncScheduleCond) {
	wg := &sync.WaitGroup{}
	currentScheduleRefIdMap := map[int]struct{}{}
	for _, schedule := range allRecords {
		wg.Add(1)
		s := schedule
		srv.executorPool.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					srv.logger.Error(ctx, "scheduleService syncSchedules syncSchedule panic", nil,
						zap.Any(logger.PanicMessage, r),
						zap.Any("schedule", s),
					)
				}
				wg.Done()
			}()

			if err := srv.syncSchedule(ctx, s); err != nil {
				srv.logger.Error(ctx, "scheduleService syncSchedules syncSchedule", err, zap.String("record_id", s.Id.Value))
			}
		})

		scheduleRefId, err := s.Id.ToId()
		if err != nil {
			srv.logger.Error(ctx, "scheduleService syncSchedules scheduleId ToId", err,
				zap.String("record_id", s.Id.Value),
				zap.String("teacher_name", s.TeacherName.Value),
				zap.String("class_type", s.ClassType.Value),
				zap.String("class_level", s.ClassLevel.Value),
				zap.String("class_time", s.ClassTime.Value),
			)
			continue
		}

		currentScheduleRefIdMap[scheduleRefId] = struct{}{}
	}
	wg.Wait()

	// 刪除已經不存在 kintone 的 schedules
	db := srv.db.Session()
	isDeleted := false
	poScheduleCond := &po.GetScheduleCond{
		IsDeleted: &isDeleted,
	}
	if cond != nil {
		if cond.StudentName != nil && cond.ParentPhone != nil {
			boStudent, err := srv.studentCommonSrv.GetStudent(ctx, &bo.StudentCond{StudentName: *cond.StudentName, ParentPhone: *cond.ParentPhone})
			if err != nil {
				if errors.Is(err, errs.DbErr.NoRow) {
					srv.logger.Error(ctx, "scheduleService syncSchedules studentCommonSrv.GetStudent", errs.StudentErr.StudentNotFoundErr,
						zap.String("student_name", *cond.StudentName),
						zap.String("parent_phone", *cond.ParentPhone),
					)
					return
				}

				srv.logger.Error(ctx, "scheduleService syncSchedules studentCommonSrv.GetStudent", err,
					zap.String("student_name", *cond.StudentName),
					zap.String("parent_phone", *cond.ParentPhone),
				)

				return
			}

			poScheduleCond.StudentId = boStudent.StudentId
		}
		if cond.ClassTimeStart != nil {
			poScheduleCond.ClassTimeStart = *cond.ClassTimeStart
		}
		if cond.ClassTimeEnd != nil {
			poScheduleCond.ClassTimeEnd = *cond.ClassTimeEnd
		}
	}
	allScheduleRefIdsInDb, err := srv.scheduleRepo.GetAllScheduleRefIds(ctx, db, poScheduleCond)
	if err != nil {
		srv.logger.Error(ctx, "scheduleService BatchSyncSchedule GetAllScheduleRefIds", err)
		return
	}
	for _, scheduleRefIdInDb := range allScheduleRefIdsInDb {
		if _, found := currentScheduleRefIdMap[scheduleRefIdInDb]; !found {
			if err := srv.DeleteSchedule(ctx, &bo.UpdateScheduleCond{ScheduleRefId: scheduleRefIdInDb}); err != nil {
				srv.logger.Error(ctx, "scheduleService BatchSyncSchedule DeleteSchedule", err, zap.Int("schedule_ref_id", scheduleRefIdInDb))
			}
		}
	}
}

func (srv *ScheduleService) syncSchedule(ctx context.Context, rawSchedule dto.ScheduleRecord) error {
	boKintoneSchedules, err := rawSchedule.ToSchedules()
	if err != nil {
		return xerrors.Errorf("scheduleService syncSchedule schedule.ToSchedules: %w", err)
	}
	// 這邊不用檢查 ToId error 因為在 checkBasicRequestData 中的 ToSchedules 已經有檢查過
	scheduleRefId, _ := rawSchedule.Id.ToId()

	// 1. 找出 db 所有對應這筆 schedule id 的資料
	db := srv.db.Session()
	schedulesInDb, err := srv.scheduleRepo.GetSchedulesByRefId(ctx, db, scheduleRefId)
	if err != nil && !errors.Is(err, errs.DbErr.NoRow) {
		return xerrors.Errorf("scheduleService syncSchedule scheduleSrv.GetSchedulesByRefId: %w", err)
	}

	// 失敗記錄
	failedRecords := make([]FailedRecord, 0, len(boKintoneSchedules)+len(schedulesInDb))

	schedulesInDbMap := map[bo.StudentCond]*po.ScheduleView{}
	currentRecordStudentNames := map[bo.StudentCond]struct{}{}
	for _, scheduleInDb := range schedulesInDb {
		schedulesInDbMap[bo.StudentCond{StudentName: scheduleInDb.StudentName, ParentPhone: scheduleInDb.ParentPhone}] = scheduleInDb
	}
	// 2. 新增 (kintone 有，db沒有)
	// 3. 更新 (kintone 有，db有)
	for _, kintoneSchedule := range boKintoneSchedules {
		// 新增
		if schedulesInDbMap[bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone}] == nil {
			err := srv.AddSchedule(ctx, kintoneSchedule, &bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone})
			if err != nil {
				srv.logger.Error(ctx, "scheduleService syncSchedule AddSchedule", err, zap.String("student_name", kintoneSchedule.StudentName))
				failedRecords = append(failedRecords, FailedRecord{ScheduleRefId: kintoneSchedule.ScheduleRefId, Name: kintoneSchedule.StudentName})
			}
		} else {
			// 更新
			scheduleInDb := schedulesInDbMap[bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone}]
			if err := srv.UpdateScheduleById(
				ctx,
				scheduleInDb.ScheduleId,
				&bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone},
				&bo.UpdateScheduleData{
					ClassLevel:  kintoneSchedule.ClassLevel,
					ClassType:   kintoneSchedule.ClassType,
					ClassTime:   kintoneSchedule.ClassTime,
					TeacherName: kintoneSchedule.TeacherName,
				}); err != nil {
				srv.logger.Error(ctx, "scheduleService syncSchedule UpdateScheduleById", err,
					zap.Int64("schedule_id", scheduleInDb.ScheduleId),
					zap.Int("schedule_ref_id", scheduleInDb.ScheduleRefId),
					zap.String("student_name", kintoneSchedule.StudentName),
				)
				failedRecords = append(failedRecords, FailedRecord{ScheduleId: scheduleInDb.ScheduleId, ScheduleRefId: scheduleInDb.ScheduleRefId, Name: scheduleInDb.StudentName})
				continue
			}
			//如果在 kintone 上刪除了照理說不會再拿到相同 id 的資料，這裡是為了避免程式邏輯錯誤導致誤刪除了不該刪的記錄
			//所以 API 如果取得到這筆確實存在的記錄但是 DB 卻標示為刪除時，還是重置該記錄的刪除狀態
			if scheduleInDb.IsDeleted || scheduleInDb.DeletedAt != nil {
				if err := srv.scheduleRepo.ResetFromDeleted(ctx, db, (po.Schedule{}).TableName(), func(db *gorm.DB) *gorm.DB {
					db.Where("schedule_id = ?", scheduleInDb.ScheduleId)
					return db
				}); err != nil {
					failedRecords = append(failedRecords, FailedRecord{ScheduleId: scheduleInDb.ScheduleId, ScheduleRefId: scheduleInDb.ScheduleRefId, Name: scheduleInDb.StudentName})
					return xerrors.Errorf("scheduleService syncSchedule ResetFromDeleted: %w", err)
				}
			}
		}
		// 記錄現有的 student names
		currentRecordStudentNames[bo.StudentCond{StudentName: kintoneSchedule.StudentName, ParentPhone: kintoneSchedule.ParentPhone}] = struct{}{}
	}
	// 4. 刪除 (kintone 沒有，db有)
	for _, scheduleInDb := range schedulesInDb {
		if _, found := currentRecordStudentNames[bo.StudentCond{StudentName: scheduleInDb.StudentName, ParentPhone: scheduleInDb.ParentPhone}]; !found {
			delErr := srv.DeleteSchedule(ctx, &bo.UpdateScheduleCond{ScheduleId: scheduleInDb.ScheduleId})
			if delErr != nil {
				srv.logger.Error(ctx, "scheduleService syncSchedule DeleteSchedule", err, zap.Int("schedule_ref_id", scheduleInDb.ScheduleRefId))
				failedRecords = append(failedRecords, FailedRecord{ScheduleId: scheduleInDb.ScheduleId, ScheduleRefId: scheduleInDb.ScheduleRefId, Name: scheduleInDb.StudentName})
			}
		}
	}

	if len(failedRecords) > 0 {
		srv.logger.Warn(ctx, "sync failed schedule records", zap.ObjectValues("records", failedRecords))
	}

	return nil
}
