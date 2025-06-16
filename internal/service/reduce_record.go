package service

import (
	"context"
	"errors"
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/SeanZhenggg/go-utils/snowflake/autoId"
	"github.com/panjf2000/ants"
	"go.uber.org/zap"
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

func ProvideReduceRecordService(
	ReduceRecordRepo interfaces.IReduceRecordRepo,
	DB database.IPostgresDB,
	studentCommonSrv interfaces.IStudentCommonSrv,
	logger logger.ILogger,
	reduceRecordCommonSrv interfaces.IReduceRecordCommonSrv,
) *ReduceRecordService {
	return &ReduceRecordService{
		recordRepo:            ReduceRecordRepo,
		DB:                    DB,
		studentCommonSrv:      studentCommonSrv,
		logger:                logger,
		reduceRecordCommonSrv: reduceRecordCommonSrv,
		executorPool:          pool.NewExecutorPool(100),
	}
}

type ReduceRecordService struct {
	recordRepo            interfaces.IReduceRecordRepo
	DB                    database.IPostgresDB
	studentCommonSrv      interfaces.IStudentCommonSrv
	logger                logger.ILogger
	reduceRecordCommonSrv interfaces.IReduceRecordCommonSrv
	executorPool          *ants.Pool `wire:"-"`
}

func (srv *ReduceRecordService) GetReduceRecords(ctx context.Context, cond *bo.ReduceRecordCond, studentCond *bo.StudentCond) ([]*bo.ReduceRecord, *po.PagerResult, error) {
	poReduceRecordCond := &po.ReduceRecordCond{}
	if !cond.ClassTimeStart.IsZero() {
		poReduceRecordCond.ClassTimeStart = cond.ClassTimeStart
	}
	if !cond.ClassTimeEnd.IsZero() {
		poReduceRecordCond.ClassTimeEnd = cond.ClassTimeEnd.Add(24 * time.Hour).Add(-1 * time.Minute)
	}
	if cond.IsDeleted != nil {
		poReduceRecordCond.IsDeleted = cond.IsDeleted
	} else {
		isDeleted := false
		poReduceRecordCond.IsDeleted = &isDeleted
	}

	var (
		boStudent *bo.Student
		err       error
	)
	// 有帶 student 條件就以學生為基礎撈資料
	if !reflect.ValueOf(studentCond).Elem().IsZero() {
		boStudent, err = srv.studentCommonSrv.GetStudent(ctx, studentCond)
		if err != nil {
			if errors.Is(err, errs.DbErr.NoRow) {
				return nil, nil, xerrors.Errorf("studentCommonSrv.GetStudent: %w", errs.StudentErr.StudentNotFoundErr)
			}

			return nil, nil, xerrors.Errorf("studentCommonSrv.GetStudent: %w", err)
		}

		poReduceRecordCond.StudentId = boStudent.StudentId
	}

	poPager := &po.Pager{
		Index: cond.Index,
		Size:  cond.Size,
		Order: cond.Order,
	}

	db := srv.DB.Session()
	poRecords, err := srv.recordRepo.GetRecordsWithSettleRecords(ctx, db, poReduceRecordCond, poPager)
	if err != nil {
		return nil, nil, xerrors.Errorf("recordRepo.GetRecordsWithSettleRecords: %w", err)
	}
	poPagerResult, err := srv.recordRepo.GetRecordsWithSettleRecordsPager(ctx, db, poReduceRecordCond, poPager)
	if err != nil {
		return nil, nil, xerrors.Errorf("recordRepo.GetRecordsWithSettleRecordsPager: %w", err)
	}

	boRecords := make([]*bo.ReduceRecord, 0, len(poRecords))
	for _, record := range poRecords {
		boRecord := &bo.ReduceRecord{
			RecordId:     record.RecordId,
			RecordRefId:  record.RecordRefId,
			StudentName:  record.StudentName,
			ParentPhone:  record.ParentPhone,
			ReducePoints: record.ReducePoints,
			IsDeleted:    record.IsDeleted,
			CreatedAt:    record.CreatedAt,
			UpdatedAt:    record.UpdatedAt,
			DeletedAt:    record.DeletedAt,
			RecordType:   record.RecordType,
		}
		if record.RecordType == "reduce_record" {
			boRecord.ClassLevel = kintone.ClassLevelToEnum(record.ClassLevel)
			boRecord.ClassType = kintone.ClassTypeToEnum(record.ClassType)
			boRecord.ClassTime = *record.ClassTime
			boRecord.TeacherName = record.TeacherName
			boRecord.IsAttended = record.IsAttended
		} else {
			boRecord.ClassTime = *record.ClassTime
		}

		boRecords = append(boRecords, boRecord)
	}

	return boRecords, poPagerResult, nil
}

func (srv *ReduceRecordService) AddReduceRecord(ctx context.Context, data *bo.KintoneReduceRecord, studentCond *bo.StudentCond) error {
	if len(studentCond.StudentName) == 0 || len(studentCond.ParentPhone) == 0 {
		return errs.ReduceRecordErr.InvalidStudentNameErr
	}

	boStudent, err := srv.studentCommonSrv.GetStudent(ctx, studentCond)
	if err != nil {
		if errors.Is(err, errs.DbErr.NoRow) {
			return xerrors.Errorf("studentCommonSrv.GetStudent: %w", errs.StudentErr.StudentNotFoundErr)
		}

		return xerrors.Errorf("studentCommonSrv.GetStudent: %w", err)
	}

	recordId, err := autoId.DefaultSnowFlake.GenNextId()
	if err != nil {
		return xerrors.Errorf("autoId.DefaultSnowFlake.GenNextId: %w", err)
	}

	poReduceRecord := &po.ReduceRecord{
		RecordId:     recordId,
		RecordRefId:  data.Id,
		StudentId:    boStudent.StudentId,
		ClassLevel:   data.ClassLevel.ToKey(),
		ClassType:    data.ClassType.ToKey(),
		ClassTime:    &data.ClassTime,
		TeacherName:  data.TeacherName,
		ReducePoints: data.ReducePoints,
		IsAttended:   data.AttendStatus,
	}

	db := srv.DB.Session()
	if err := srv.recordRepo.AddRecord(ctx, db, poReduceRecord); err != nil {
		return xerrors.Errorf("recordRepo.AddRecord: %w", err)
	}

	return nil
}

func (srv *ReduceRecordService) UpdateReduceRecord(ctx context.Context, cond *bo.ReduceRecordCond, studentCond *bo.StudentCond, data *bo.UpdateReduceRecordData) error {
	if len(studentCond.StudentName) == 0 || len(studentCond.ParentPhone) == 0 {
		return errs.ReduceRecordErr.InvalidStudentNameErr
	}

	boStudent, err := srv.studentCommonSrv.GetStudent(ctx, studentCond)
	if err != nil {
		if errors.Is(err, errs.DbErr.NoRow) {
			return xerrors.Errorf("ReduceRecordService UpdateReduceRecord studentCommonSrv.GetStudent: %w", errs.StudentErr.StudentNotFoundErr)
		}

		return xerrors.Errorf("ReduceRecordService UpdateReduceRecord studentCommonSrv.GetStudent: %w", err)
	}

	poReduceRecordCond := &po.ReduceRecordCond{RecordRefId: cond.RecordRefId}
	db := srv.DB.Session()
	poUpdateReduceRecordData := &po.UpdateReduceRecordData{
		StudentId:    boStudent.StudentId,
		ClassType:    data.ClassType.ToKey(),
		ClassLevel:   data.ClassLevel.ToKey(),
		ClassTime:    &data.ClassTime,
		TeacherName:  data.TeacherName,
		ReducePoints: data.ReducePoints,
		IsAttended:   data.IsAttended,
	}

	if err := srv.recordRepo.UpdateRecord(ctx, db, poReduceRecordCond, poUpdateReduceRecordData); err != nil {
		return xerrors.Errorf("reduceRecordService UpdateReduceRecord recordRepo.UpdateRecord: %w", err)
	}

	return nil
}

func (srv *ReduceRecordService) DeleteReduceRecord(ctx context.Context, cond *bo.ReduceRecordCond) error {
	deleted := true
	poReduceRecordCond := &po.ReduceRecordCond{
		RecordRefId: cond.RecordRefId,
	}
	now := time.Now()
	poUpdateReduceRecordData := &po.UpdateReduceRecordData{
		IsDeleted: &deleted,
		DeletedAt: &now,
	}

	db := srv.DB.Session()
	if err := srv.recordRepo.UpdateRecord(ctx, db, poReduceRecordCond, poUpdateReduceRecordData); err != nil {
		return xerrors.Errorf("reduceRecordService DeleteReduceRecord recordRepo.UpdateRecord: %w", err)
	}

	return nil
}

func (srv *ReduceRecordService) BatchSyncReduceRecord(ctx context.Context, cond *bo.SyncReduceRecordCond, wg ...*sync.WaitGroup) error {
	boReduceRecordReq := &dto.ReduceRecordReq{
		Limit:  request.GetRecordsBatchLimit,
		Offset: request.GetRecordsBatchOffset,
	}

	var (
		total                int
		err                  error
		kintoneReduceRecords []*bo.KintoneReduceRecord
	)

	if cond != nil {
		if cond.StudentName != nil && cond.ParentPhone != nil {
			boReduceRecordReq.StudentName = strUtil.GetFullStudentName(*cond.StudentName, *cond.ParentPhone)
		}
		if cond.Limit != nil {
			boReduceRecordReq.Limit = *cond.Limit
		}
		if cond.Offset != nil {
			boReduceRecordReq.Offset = *cond.Offset
		}
		if cond.ClassTimeStart != nil {
			boReduceRecordReq.ClassTimeStart = *cond.ClassTimeStart
		}
		if cond.ClassTimeEnd != nil {
			boReduceRecordReq.ClassTimeEnd = *cond.ClassTimeEnd
		}
	}

	kintoneReduceRecords, total, err = srv.reduceRecordCommonSrv.GetKintoneReduceRecords(ctx, boReduceRecordReq)
	if err != nil {
		return xerrors.Errorf("reduceRecordService BatchSyncReduceRecord GetKintoneReduceRecord: %w", err)
	}

	allRecords := make([]*bo.KintoneReduceRecord, 0, total)
	allRecords = append(allRecords, kintoneReduceRecords...)

	offset, limit := boReduceRecordReq.Offset, boReduceRecordReq.Limit
	for ((offset * limit) + limit) < total {
		offset += 1
		boReduceRecordReq.Offset = offset * limit
		kintoneReduceRecords, _, err = srv.reduceRecordCommonSrv.GetKintoneReduceRecords(ctx, boReduceRecordReq)
		if err != nil {
			return xerrors.Errorf("reduceRecordService BatchSyncReduceRecord GetKintoneReduceRecord: %w", err)
		}
		allRecords = append(allRecords, kintoneReduceRecords...)
	}

	var wait *sync.WaitGroup
	if len(wg) > 0 {
		wait = wg[0]
		wait.Add(1)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				srv.logger.Error(ctx, "reduceRecordService syncReduceRecords panic", nil, zap.Any(logger.PanicMessage, r))
			}
			if wait != nil {
				wait.Done()
			}
		}()

		srv.syncReduceRecords(ctx, allRecords, cond)
	}()

	return nil
}

func (srv *ReduceRecordService) GetStudentTotalReducePoints(ctx context.Context, db *gorm.DB, cond *bo.StudentTotalReducePointsCond) (map[int64]*bo.StudentTotalReducePoints, error) {
	if len(cond.StudentIds) == 0 {
		return nil, nil
	}

	poCond := &po.ReduceRecordCond{StudentIds: cond.StudentIds}

	if !cond.ClassTimeStart.IsZero() {
		poCond.ClassTimeStart = cond.ClassTimeStart
	}
	if !cond.ClassTimeEnd.IsZero() {
		poCond.ClassTimeEnd = cond.ClassTimeEnd
	}

	isDeleted := false
	poCond.IsDeleted = &isDeleted

	studentTotalReducePoints, err := srv.recordRepo.GetStudentTotalReducePoints(ctx, db, poCond)
	if err != nil {
		return nil, xerrors.Errorf("recordRepo.GetStudentTotalReducePoints: %w", err)
	}

	boStudentTotalReducePoints := make(map[int64]*bo.StudentTotalReducePoints, len(studentTotalReducePoints))
	for _, studentTotalDepositPoint := range studentTotalReducePoints {
		boStudentTotalReducePoints[studentTotalDepositPoint.StudentId] = &bo.StudentTotalReducePoints{
			StudentId:         studentTotalDepositPoint.StudentId,
			TotalReducePoints: studentTotalDepositPoint.TotalReducePoints,
		}
	}

	return boStudentTotalReducePoints, nil
}

func (srv *ReduceRecordService) syncReduceRecords(ctx context.Context, allRecords []*bo.KintoneReduceRecord, cond *bo.SyncReduceRecordCond) {
	wg := &sync.WaitGroup{}
	currentRecordRefIdMap := map[int]struct{}{}
	for _, reduceRecord := range allRecords {
		wg.Add(1)
		rr := reduceRecord
		srv.executorPool.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					srv.logger.Error(ctx, "reduceRecordService syncReduceRecords syncReduceRecord panic", nil,
						zap.Any(logger.PanicMessage, r),
						zap.Any("reduceRecord", *rr),
					)
				}
				wg.Done()
			}()
			if err := srv.syncReduceRecord(ctx, rr); err != nil {
				srv.logger.Error(ctx, "reduceRecordService syncReduceRecords syncReduceRecord", err, zap.Int("record_ref_id", rr.Id), zap.String("student_name", rr.KintoneStudentName))
			}
		})

		currentRecordRefIdMap[rr.Id] = struct{}{}
	}

	wg.Wait()

	db := srv.DB.Session()
	isDeleted := false
	poReduceRecordCond := &po.ReduceRecordCond{
		IsDeleted: &isDeleted,
	}
	if cond != nil {
		if cond.StudentName != nil && cond.ParentPhone != nil {
			boStudent, err := srv.studentCommonSrv.GetStudent(ctx, &bo.StudentCond{StudentName: *cond.StudentName, ParentPhone: *cond.ParentPhone})
			if err != nil {
				if errors.Is(err, errs.DbErr.NoRow) {
					srv.logger.Error(ctx, "reduceRecordService syncReduceRecords studentCommonSrv.GetStudent", errs.StudentErr.StudentNotFoundErr,
						zap.String("student_name", *cond.StudentName),
						zap.String("parent_phone", *cond.ParentPhone),
					)
					return
				}

				srv.logger.Error(ctx, "reduceRecordService syncReduceRecords studentCommonSrv.GetStudent", err,
					zap.String("student_name", *cond.StudentName),
					zap.String("parent_phone", *cond.ParentPhone),
				)

				return
			}

			poReduceRecordCond.StudentId = boStudent.StudentId
		}
		if cond.ClassTimeStart != nil {
			poReduceRecordCond.ClassTimeStart = *cond.ClassTimeStart
		}
		if cond.ClassTimeEnd != nil {
			poReduceRecordCond.ClassTimeEnd = *cond.ClassTimeEnd
		}
	}
	recordRefIds, err := srv.recordRepo.GetReduceRecordRefIds(ctx, db, poReduceRecordCond)
	if err != nil {
		srv.logger.Error(ctx, "reduceRecordService syncReduceRecords GetReduceRecordRefIds", err)
		return
	}

	for _, recordRefId := range recordRefIds {
		if _, found := currentRecordRefIdMap[recordRefId]; !found {
			if err := srv.DeleteReduceRecord(ctx, &bo.ReduceRecordCond{RecordRefId: recordRefId}); err != nil {
				srv.logger.Error(ctx, "reduceRecordService syncReduceRecords DeleteReduceRecord", err, zap.Int("record_ref_id", recordRefId))
			}
		}
	}
}

func (srv *ReduceRecordService) syncReduceRecord(ctx context.Context, data *bo.KintoneReduceRecord) error {
	db := srv.DB.Session()
	poReduceRecordCond := &po.ReduceRecordCond{RecordRefId: data.Id}
	record, err := srv.recordRepo.GetRecord(ctx, db, poReduceRecordCond)
	if err != nil && !errors.Is(err, errs.DbErr.NoRow) {
		return xerrors.Errorf("reduceRecordService syncReduceRecord recordRepo.GetRecord: %w", err)
	}

	boStudentCond := &bo.StudentCond{StudentName: data.StudentName, ParentPhone: data.ParentPhone}
	if record != nil {
		// Update
		boReduceRecordCond := &bo.ReduceRecordCond{RecordRefId: data.Id}
		boUpdateReduceRecordData := &bo.UpdateReduceRecordData{
			ClassLevel:   data.ClassLevel,
			ClassType:    data.ClassType,
			ClassTime:    data.ClassTime,
			TeacherName:  data.TeacherName,
			ReducePoints: &data.ReducePoints,
			IsAttended:   &data.AttendStatus,
		}
		if err := srv.UpdateReduceRecord(ctx, boReduceRecordCond, boStudentCond, boUpdateReduceRecordData); err != nil {
			return xerrors.Errorf("reduceRecordService syncReduceRecord UpdateReduceRecord: %w", err)
		}
		//如果在 kintone 上刪除了照理說不會再拿到相同 id 的資料，這裡是為了避免程式邏輯錯誤導致誤刪除了不該刪的記錄
		//所以 API 如果取得到這筆確實存在的記錄但是 DB 卻標示為刪除時，還是重置該記錄的刪除狀態
		if record.IsDeleted || record.DeletedAt != nil {
			if err := srv.recordRepo.ResetFromDeleted(ctx, db, (po.ReduceRecord{}).TableName(), func(db *gorm.DB) *gorm.DB {
				db.Where("record_id = ?", record.RecordId)
				return db
			}); err != nil {
				return xerrors.Errorf("reduceRecordService syncReduceRecord ResetFromDeleted: %w", err)
			}
		}
	} else {
		// Create
		if err := srv.AddReduceRecord(ctx, data, boStudentCond); err != nil {
			return xerrors.Errorf("reduceRecordService syncReduceRecord AddReduceRecord: %w", err)
		}
	}

	return nil
}
