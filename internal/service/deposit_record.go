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

func ProvideDepositRecordService(
	depositRecordRepo interfaces.IDepositRecordRepo,
	DB database.IPostgresDB,
	studentCommonSrv interfaces.IStudentCommonSrv,
	logger logger.ILogger,
	depositRecordCommonSrv interfaces.IDepositRecordCommonSrv,
) *DepositRecordService {
	return &DepositRecordService{
		recordRepo:             depositRecordRepo,
		DB:                     DB,
		studentCommonSrv:       studentCommonSrv,
		logger:                 logger,
		depositRecordCommonSrv: depositRecordCommonSrv,
		executorPool:           pool.NewExecutorPool(50),
	}
}

type DepositRecordService struct {
	recordRepo             interfaces.IDepositRecordRepo
	DB                     database.IPostgresDB
	studentCommonSrv       interfaces.IStudentCommonSrv
	logger                 logger.ILogger
	depositRecordCommonSrv interfaces.IDepositRecordCommonSrv
	executorPool           *ants.Pool `wire:"-"`
}

func (srv *DepositRecordService) GetDepositRecords(ctx context.Context, cond *bo.DepositRecordCond, studentCond *bo.StudentCond) ([]*bo.DepositRecord, *po.PagerResult, error) {
	poDepositRecordCond := &po.DepositRecordCond{}
	if !cond.ChargingDateStart.IsZero() {
		poDepositRecordCond.ChargingDateStart = cond.ChargingDateStart
	}
	if !cond.ChargingDateEnd.IsZero() {
		poDepositRecordCond.ChargingDateEnd = cond.ChargingDateEnd.Add(24 * time.Hour).Add(-1 * time.Minute)
	}
	if cond.IsDeleted != nil {
		poDepositRecordCond.IsDeleted = cond.IsDeleted
	} else {
		isDeleted := false
		poDepositRecordCond.IsDeleted = &isDeleted
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
				return nil, nil, xerrors.Errorf("depositRecordService GetDepositRecords studentCommonSrv.GetStudent: %w", errs.StudentErr.StudentNotFoundErr)
			}
			return nil, nil, xerrors.Errorf("depositRecordService GetDepositRecords studentCommonSrv.GetStudent: %w", err)
		}

		poDepositRecordCond.StudentId = boStudent.StudentId
	}

	poPager := &po.Pager{
		Index: cond.Index,
		Size:  cond.Size,
		Order: cond.Order,
	}

	db := srv.DB.Session()
	poRecords, err := srv.recordRepo.GetRecords(ctx, db, poDepositRecordCond, poPager)
	if err != nil {
		return nil, nil, xerrors.Errorf("depositRecordService GetDepositRecords recordRepo.GetRecords: %w", err)
	}
	poPagerResult, err := srv.recordRepo.GetRecordsPager(ctx, db, poDepositRecordCond, poPager)
	if err != nil {
		return nil, nil, xerrors.Errorf("depositRecordService GetDepositRecords recordRepo.GetRecordsPager: %w", err)
	}

	boRecords := make([]*bo.DepositRecord, 0, len(poRecords))
	for _, record := range poRecords {
		boDepositRecord := &bo.DepositRecord{
			RecordId:             record.RecordId,
			RecordRefId:          record.RecordRefId,
			StudentName:          record.StudentName,
			ParentPhone:          record.ParentPhone,
			ChargingDate:         *record.ChargingDate,
			TaxId:                record.TaxId,
			AccountLastFiveYards: record.AccountLastFiveYards,
			ChargingAmount:       record.ChargingAmount,
			TeacherName:          record.TeacherName,
			DepositedPoints:      record.DepositedPoints,
			ChargingMethod:       kintone.ChargingMethodToEnum(record.ChargingMethod.Values),
			ActualChargingAmount: record.ActualChargingAmount,
			IsDeleted:            record.IsDeleted,
			CreatedAt:            record.CreatedAt,
			UpdatedAt:            record.UpdatedAt,
			DeletedAt:            record.DeletedAt,
		}
		if record.HitStatus {
			boDepositRecord.ChargingStatus = kintone.IsCharged
		}
		boRecords = append(boRecords, boDepositRecord)
	}

	return boRecords, poPagerResult, nil
}

func (srv *DepositRecordService) AddDepositRecord(ctx context.Context, data *bo.KintoneDepositRecord, studentCond *bo.StudentCond) error {
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

	poDepositRecord := &po.DepositRecord{
		RecordId:             recordId,
		RecordRefId:          data.Id,
		StudentId:            boStudent.StudentId,
		ChargingDate:         &data.ChargingDate,
		TaxId:                data.TaxId,
		AccountLastFiveYards: data.AccountLastFiveYards,
		ChargingAmount:       data.ChargingAmount,
		TeacherName:          data.TeacherName,
		DepositedPoints:      data.DepositedPoints,
		ChargingMethod:       po.ChargingMethodArray{Values: kintone.ChargingMethodToKey(data.ChargingMethod)},
		HitStatus:            data.ChargingStatus.ToValue(),
		ActualChargingAmount: data.ActualChargingAmount,
	}

	db := srv.DB.Session()
	if err := srv.recordRepo.AddRecord(ctx, db, poDepositRecord); err != nil {
		return xerrors.Errorf("recordRepo.AddRecord: %w", err)
	}

	return nil
}

func (srv *DepositRecordService) UpdateDepositRecord(ctx context.Context, cond *bo.DepositRecordCond, studentCond *bo.StudentCond, data *bo.UpdateDepositRecordData) error {
	boStudent, err := srv.studentCommonSrv.GetStudent(ctx, studentCond)
	if err != nil {
		if errors.Is(err, errs.DbErr.NoRow) {
			return xerrors.Errorf("depositRecordService UpdateDepositRecord studentCommonSrv.GetStudent: %w", errs.StudentErr.StudentNotFoundErr)
		}

		return xerrors.Errorf("depositRecordService UpdateDepositRecord studentCommonSrv.GetStudent: %w", err)
	}

	db := srv.DB.Session()
	poDepositRecordCond := &po.DepositRecordCond{RecordRefId: cond.RecordRefId}
	hitStatus := data.ChargingStatus.ToValue()
	poUpdateDepositRecordData := &po.UpdateDepositRecordData{
		StudentId:            boStudent.StudentId,
		ChargingDate:         &data.ChargingDate,
		TaxId:                data.TaxId,
		AccountLastFiveYards: data.AccountLastFiveYards,
		ChargingAmount:       data.ChargingAmount,
		TeacherName:          data.TeacherName,
		DepositedPoints:      data.DepositedPoints,
		ChargingMethod:       po.ChargingMethodArray{Values: kintone.ChargingMethodToKey(data.ChargingMethod)},
		HitStatus:            &hitStatus,
		ActualChargingAmount: data.ActualChargingAmount,
	}
	if err := srv.recordRepo.UpdateRecord(ctx, db, poDepositRecordCond, poUpdateDepositRecordData); err != nil {
		return xerrors.Errorf("depositRecordService UpdateDepositRecord recordRepo.UpdateRecord: %w", err)
	}

	return nil
}

func (srv *DepositRecordService) DeleteDepositRecord(ctx context.Context, cond *bo.DepositRecordCond) error {
	deleted := true
	poDepositRecordCond := &po.DepositRecordCond{
		RecordRefId: cond.RecordRefId,
	}
	now := time.Now()
	poUpdateDepositRecordData := &po.UpdateDepositRecordData{
		IsDeleted: &deleted,
		DeletedAt: &now,
	}

	db := srv.DB.Session()
	if err := srv.recordRepo.UpdateRecord(ctx, db, poDepositRecordCond, poUpdateDepositRecordData); err != nil {
		return xerrors.Errorf("depositRecordService DeleteDepositRecord recordRepo.UpdateRecord: %w", err)
	}

	return nil
}

func (srv *DepositRecordService) BatchSyncDepositRecord(ctx context.Context, cond *bo.SyncDepositRecordCond, wg ...*sync.WaitGroup) error {
	boDepositRecordReq := &dto.DepositRecordReq{}

	var (
		total                 int
		err                   error
		kintoneDepositRecords []*bo.KintoneDepositRecord
	)

	boDepositRecordReq.Limit = request.GetRecordsBatchLimit
	boDepositRecordReq.Offset = request.GetRecordsBatchOffset

	if cond != nil {
		if cond.StudentName != nil && cond.ParentPhone != nil {
			boDepositRecordReq.StudentName = strUtil.GetFullStudentName(*cond.StudentName, *cond.ParentPhone)
		}
		if cond.Limit != nil {
			boDepositRecordReq.Limit = *cond.Limit
		}
		if cond.Offset != nil {
			boDepositRecordReq.Offset = *cond.Offset
		}
		if cond.ChargingDateStart != nil {
			boDepositRecordReq.ChargingDateStart = *cond.ChargingDateStart
		}
		if cond.ChargingDateEnd != nil {
			boDepositRecordReq.ChargingDateEnd = *cond.ChargingDateEnd
		}
	}

	kintoneDepositRecords, total, err = srv.depositRecordCommonSrv.GetKintoneDepositRecords(ctx, boDepositRecordReq)
	if err != nil {
		return xerrors.Errorf("depositRecordService BatchSyncDepositRecord GetKintoneDepositRecord: %w", err)
	}

	allRecords := make([]*bo.KintoneDepositRecord, 0, total)
	allRecords = append(allRecords, kintoneDepositRecords...)

	offset, limit := boDepositRecordReq.Offset, boDepositRecordReq.Limit
	for ((offset * limit) + limit) < total {
		offset += 1
		boDepositRecordReq.Offset = offset * limit
		kintoneDepositRecords, _, err = srv.depositRecordCommonSrv.GetKintoneDepositRecords(ctx, boDepositRecordReq)
		if err != nil {
			return xerrors.Errorf("depositRecordService BatchSyncDepositRecord GetKintoneDepositRecord: %w", err)
		}
		allRecords = append(allRecords, kintoneDepositRecords...)
	}

	var wait *sync.WaitGroup
	if len(wg) > 0 {
		wait = wg[0]
		wait.Add(1)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				srv.logger.Error(ctx, "depositRecordService syncDepositRecords panic", nil,
					zap.Any(logger.PanicMessage, r),
				)
			}
			if wait != nil {
				wait.Done()
			}
		}()

		srv.syncDepositRecords(ctx, allRecords, cond)
	}()

	return nil
}

func (srv *DepositRecordService) GetStudentTotalDepositPoints(ctx context.Context, db *gorm.DB, cond *bo.StudentTotalDepositPointsCond) (map[int64]*bo.StudentTotalDepositPoints, error) {
	if len(cond.StudentIds) == 0 {
		return nil, errors.New("invalid parameter StudentIds")
	}

	poCond := &po.DepositRecordCond{StudentIds: cond.StudentIds}

	if !cond.ChargingDateStart.IsZero() {
		poCond.ChargingDateStart = cond.ChargingDateStart
	}
	if !cond.ChargingDateEnd.IsZero() {
		poCond.ChargingDateEnd = cond.ChargingDateEnd
	}

	isDeleted := false
	poCond.IsDeleted = &isDeleted

	studentTotalDepositPoints, err := srv.recordRepo.GetStudentTotalDepositPoints(ctx, db, poCond)
	if err != nil {
		return nil, xerrors.Errorf("recordRepo.GetStudentTotalDepositPoints: %w", err)
	}

	boStudentTotalDepositPoints := make(map[int64]*bo.StudentTotalDepositPoints, len(studentTotalDepositPoints))
	for _, studentTotalDepositPoint := range studentTotalDepositPoints {
		boStudentTotalDepositPoints[studentTotalDepositPoint.StudentId] = &bo.StudentTotalDepositPoints{
			StudentId:          studentTotalDepositPoint.StudentId,
			TotalDepositPoints: studentTotalDepositPoint.TotalDepositedPoints,
		}
	}

	return boStudentTotalDepositPoints, nil
}

func (srv *DepositRecordService) syncDepositRecords(ctx context.Context, allRecords []*bo.KintoneDepositRecord, cond *bo.SyncDepositRecordCond) {
	wg := &sync.WaitGroup{}
	currentRecordIds := map[int]struct{}{}
	for _, kintoneDepositRecord := range allRecords {
		wg.Add(1)
		dr := kintoneDepositRecord
		srv.executorPool.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					srv.logger.Error(ctx, "depositRecordService syncDepositRecords syncDepositRecord panic", nil,
						zap.Any(logger.PanicMessage, r),
						zap.Any("depositRecord", *dr),
					)
				}
				wg.Done()
			}()

			err := srv.syncDepositRecord(ctx, dr)
			if err != nil {
				srv.logger.Error(ctx, "depositRecordService syncDepositRecords syncDepositRecord", err, zap.Int("record_ref_id", dr.Id), zap.String("student_name", dr.KintoneStudentName))
			}
		})

		currentRecordIds[dr.Id] = struct{}{}
	}

	wg.Wait()

	db := srv.DB.Session()
	isDeleted := false
	poDepositRecordCond := &po.DepositRecordCond{
		IsDeleted: &isDeleted,
	}
	if cond != nil {
		if cond.StudentName != nil && cond.ParentPhone != nil {
			boStudent, err := srv.studentCommonSrv.GetStudent(ctx, &bo.StudentCond{StudentName: *cond.StudentName, ParentPhone: *cond.ParentPhone})
			if err != nil {
				if errors.Is(err, errs.DbErr.NoRow) {
					srv.logger.Error(ctx, "depositRecordService syncDepositRecords studentCommonSrv.GetStudent", errs.StudentErr.StudentNotFoundErr,
						zap.String("student_name", *cond.StudentName),
						zap.String("parent_phone", *cond.ParentPhone),
					)
					return
				}

				srv.logger.Error(ctx, "depositRecordService syncDepositRecords studentCommonSrv.GetStudent", err,
					zap.String("student_name", *cond.StudentName),
					zap.String("parent_phone", *cond.ParentPhone),
				)

				return
			}

			poDepositRecordCond.StudentId = boStudent.StudentId
		}
		if cond.ChargingDateStart != nil {
			poDepositRecordCond.ChargingDateStart = *cond.ChargingDateStart
		}
		if cond.ChargingDateEnd != nil {
			poDepositRecordCond.ChargingDateEnd = *cond.ChargingDateEnd
		}
	}
	recordRefIds, err := srv.recordRepo.GetDepositRecordRefIds(ctx, db, poDepositRecordCond)
	if err != nil {
		srv.logger.Error(ctx, "depositRecordService syncDepositRecords GetDepositRecordRefIds", err)
		return
	}

	for _, recordRefId := range recordRefIds {
		if _, found := currentRecordIds[recordRefId]; !found {
			if err := srv.DeleteDepositRecord(ctx, &bo.DepositRecordCond{RecordRefId: recordRefId}); err != nil {
				srv.logger.Error(ctx, "depositRecordService syncDepositRecords DeleteDepositRecord", err, zap.Int("record_ref_id", recordRefId))
			}
		}
	}
}

func (srv *DepositRecordService) syncDepositRecord(ctx context.Context, data *bo.KintoneDepositRecord) error {
	db := srv.DB.Session()
	poDepositRecordCond := &po.DepositRecordCond{RecordRefId: data.Id}
	record, err := srv.recordRepo.GetRecord(ctx, db, poDepositRecordCond)
	if err != nil && !errors.Is(err, errs.DbErr.NoRow) {
		return xerrors.Errorf("depositRecordService syncDepositRecord recordRepo.GetRecord: %w", err)
	}

	boStudentCond := &bo.StudentCond{StudentName: data.StudentName, ParentPhone: data.ParentPhone}
	if record != nil {
		// Update
		boDepositRecordCond := &bo.DepositRecordCond{RecordRefId: data.Id}
		boUpdateDepositRecordData := &bo.UpdateDepositRecordData{
			ChargingDate:         data.ChargingDate,
			TaxId:                data.TaxId,
			AccountLastFiveYards: data.AccountLastFiveYards,
			ChargingAmount:       &data.ChargingAmount,
			TeacherName:          data.TeacherName,
			DepositedPoints:      &data.DepositedPoints,
			ChargingMethod:       data.ChargingMethod,
			ChargingStatus:       data.ChargingStatus,
			ActualChargingAmount: &data.ActualChargingAmount,
		}
		if err := srv.UpdateDepositRecord(ctx, boDepositRecordCond, boStudentCond, boUpdateDepositRecordData); err != nil {
			return xerrors.Errorf("depositRecordService syncDepositRecord UpdateDepositRecord: %w", err)
		}
		//如果在 kintone 上刪除了照理說不會再拿到相同 id 的資料，這裡是為了避免程式邏輯錯誤導致誤刪除了不該刪的記錄
		//所以 API 如果取得到這筆確實存在的記錄但是 DB 卻標示為刪除時，還是重置該記錄的刪除狀態
		if record.IsDeleted || record.DeletedAt != nil {
			if err := srv.recordRepo.ResetFromDeleted(ctx, db, (po.DepositRecord{}).TableName(), func(db *gorm.DB) *gorm.DB {
				db.Where("record_id = ?", record.RecordId)
				return db
			}); err != nil {
				return xerrors.Errorf("depositRecordService syncDepositRecord ResetFromDeleted: %w", err)
			}
		}
	} else {
		// Create
		if err := srv.AddDepositRecord(ctx, data, boStudentCond); err != nil {
			return xerrors.Errorf("depositRecordService syncDepositRecord AddDepositRecord: %w", err)
		}
	}

	return nil
}
