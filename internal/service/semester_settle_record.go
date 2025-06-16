package service

import (
	"context"
	"errors"
	"fmt"
	pkgLogger "github.com/SeanZhenggg/go-utils/logger"
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
	"jaystar/internal/utils"
	"jaystar/internal/utils/errs"
	"jaystar/internal/utils/pool"
	"jaystar/internal/utils/strUtil"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"time"
)

type SemesterSettleRecordService struct {
	DB                              database.IPostgresDB
	logger                          pkgLogger.ILogger
	depositRecordSrv                interfaces.IDepositRecordSrv
	reduceRecordSrv                 interfaces.IReduceRecordSrv
	kintoneSemesterSettleRecordRepo interfaces.IKintoneSemesterSettleRecordRepo
	semesterSettleRecordRepo        interfaces.ISemesterSettleRecordRepo
	studentSrv                      interfaces.IStudentSrv
	studentCommonSrv                interfaces.IStudentCommonSrv
	pointCardSrv                    interfaces.IPointCardSrv
	executorPool                    *ants.Pool `wire:"-"`
}

func ProvideSemesterSettleRecordService(
	db database.IPostgresDB,
	logger pkgLogger.ILogger,
	depositRecordSrv interfaces.IDepositRecordSrv,
	reduceRecordSrv interfaces.IReduceRecordSrv,
	kintoneSemesterSettleRecordRepo interfaces.IKintoneSemesterSettleRecordRepo,
	semesterSettleRecordRepo interfaces.ISemesterSettleRecordRepo,
	studentSrv interfaces.IStudentSrv,
	studentCommonSrv interfaces.IStudentCommonSrv,
	pointCardSrv interfaces.IPointCardSrv,
) *SemesterSettleRecordService {
	return &SemesterSettleRecordService{
		DB:                              db,
		logger:                          logger,
		depositRecordSrv:                depositRecordSrv,
		reduceRecordSrv:                 reduceRecordSrv,
		kintoneSemesterSettleRecordRepo: kintoneSemesterSettleRecordRepo,
		semesterSettleRecordRepo:        semesterSettleRecordRepo,
		studentSrv:                      studentSrv,
		studentCommonSrv:                studentCommonSrv,
		pointCardSrv:                    pointCardSrv,
		executorPool:                    pool.NewExecutorPool(30),
	}
}

var (
	semesterStartAtDates = []string{"03/01", "09/01"}
)

type settlementItem struct {
	student     *bo.Student
	clearPoints float64
}

type dateRange struct {
	Start time.Time
	End   time.Time
}

func (srv *SemesterSettleRecordService) GetSemesterSettleRecords(ctx context.Context, cond *bo.SemesterSettleRecordCond, studentCond *bo.StudentCond) ([]*bo.SemesterSettleRecord, *po.PagerResult, error) {
	poSemesterSettleRecordCond := &po.SemesterSettleRecordCond{}
	if !cond.StartTime.IsZero() {
		poSemesterSettleRecordCond.StartTime = cond.StartTime
	}
	if !cond.EndTime.IsZero() {
		poSemesterSettleRecordCond.EndTime = cond.EndTime.Add(24 * time.Hour).Add(-1 * time.Minute)
	}
	if cond.IsDeleted != nil {
		poSemesterSettleRecordCond.IsDeleted = cond.IsDeleted
	} else {
		isDeleted := false
		poSemesterSettleRecordCond.IsDeleted = &isDeleted
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

		poSemesterSettleRecordCond.StudentId = boStudent.StudentId
	}

	poPager := &po.Pager{
		Index: cond.Index,
		Size:  cond.Size,
		Order: cond.Order,
	}

	db := srv.DB.Session()
	poRecords, err := srv.semesterSettleRecordRepo.GetRecords(ctx, db, poSemesterSettleRecordCond, poPager)
	if err != nil {
		return nil, nil, xerrors.Errorf("recordRepo.GetRecords: %w", err)
	}
	poPagerResult, err := srv.semesterSettleRecordRepo.GetRecordsPager(ctx, db, poSemesterSettleRecordCond, poPager)
	if err != nil {
		return nil, nil, xerrors.Errorf("recordRepo.GetRecordsPager: %w", err)
	}

	boRecords := make([]*bo.SemesterSettleRecord, 0, len(poRecords))
	for _, record := range poRecords {
		boRecords = append(boRecords, &bo.SemesterSettleRecord{
			RecordId:    record.RecordId,
			RecordRefId: record.RecordRefId,
			StudentName: record.StudentName,
			ParentPhone: record.ParentPhone,
			StartTime:   record.StartTime,
			EndTime:     record.EndTime,
			ClearPoints: record.ClearPoints,
			IsDeleted:   record.IsDeleted,
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   record.UpdatedAt,
			DeletedAt:   record.DeletedAt,
		})
	}

	return boRecords, poPagerResult, nil
}

func (srv *SemesterSettleRecordService) AddSemesterSettleRecord(ctx context.Context, data *bo.SemesterSettleRecord, studentCond *bo.StudentCond) error {
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

	poSemesterSettleRecord := &po.SemesterSettleRecord{
		RecordId:    recordId,
		RecordRefId: data.RecordRefId,
		StudentId:   boStudent.StudentId,
		StartTime:   data.StartTime,
		EndTime:     data.EndTime,
		ClearPoints: data.ClearPoints,
	}

	db := srv.DB.Session()
	if err := srv.semesterSettleRecordRepo.AddRecord(ctx, db, poSemesterSettleRecord); err != nil {
		return xerrors.Errorf("recordRepo.AddRecord: %w", err)
	}

	return nil
}

func (srv *SemesterSettleRecordService) UpdateSemesterSettleRecord(ctx context.Context, cond *bo.UpdateSemesterSettleRecordCond, studentCond *bo.StudentCond, data *bo.UpdateSemesterSettleRecordData) error {
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

	poUpdateSemesterSettleRecordCond := &po.UpdateSemesterSettleRecordCond{RecordRefId: cond.RecordRefId}
	db := srv.DB.Session()
	poUpdateSemesterSettleRecordData := &po.UpdateSemesterSettleRecordData{
		StudentId:   &boStudent.StudentId,
		StartTime:   &data.StartTime,
		EndTime:     &data.EndTime,
		ClearPoints: data.ClearPoints,
	}

	if err := srv.semesterSettleRecordRepo.UpdateRecord(ctx, db, poUpdateSemesterSettleRecordCond, poUpdateSemesterSettleRecordData); err != nil {
		return xerrors.Errorf("recordRepo.UpdateRecord: %w", err)
	}

	return nil
}

func (srv *SemesterSettleRecordService) DeleteSemesterSettleRecord(ctx context.Context, cond *bo.UpdateSemesterSettleRecordCond) error {
	deleted := true
	poUpdateSemesterSettleRecordCond := &po.UpdateSemesterSettleRecordCond{
		RecordRefId: cond.RecordRefId,
	}
	now := time.Now()
	poUpdateSemesterSettleRecordData := &po.UpdateSemesterSettleRecordData{
		IsDeleted: &deleted,
		DeletedAt: &now,
	}

	db := srv.DB.Session()
	if err := srv.semesterSettleRecordRepo.UpdateRecord(ctx, db, poUpdateSemesterSettleRecordCond, poUpdateSemesterSettleRecordData); err != nil {
		return xerrors.Errorf("recordRepo.UpdateRecord: %w", err)
	}

	return nil
}

func (srv *SemesterSettleRecordService) BatchSyncSemesterSettleRecord(ctx context.Context, cond *bo.SyncSemesterSettleRecordCond, wg ...*sync.WaitGroup) error {
	semesterSettleRecordReq := &dto.SemesterSettleRecordReq{
		Limit:  request.GetRecordsBatchLimit,
		Offset: request.GetRecordsBatchOffset,
	}

	if cond != nil {
		if cond.StudentName != nil && cond.ParentPhone != nil {
			semesterSettleRecordReq.StudentName = strUtil.GetFullStudentName(*cond.StudentName, *cond.ParentPhone)
		}
		if cond.Limit != nil {
			semesterSettleRecordReq.Limit = *cond.Limit
		}
		if cond.Offset != nil {
			semesterSettleRecordReq.Offset = *cond.Offset
		}
		if cond.StartTime != nil {
			semesterSettleRecordReq.StartTime = *cond.StartTime
		}
		if cond.EndTime != nil {
			semesterSettleRecordReq.EndTime = *cond.EndTime
		}
	}

	allRecords, err := srv.getAllKintoneSemesterSettleRecords(ctx, semesterSettleRecordReq)
	if err != nil {
		return xerrors.Errorf("getAllKintoneSemesterSettleRecords: %w", err)
	}

	var wait *sync.WaitGroup
	if len(wg) > 0 {
		wait = wg[0]
		wait.Add(1)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				srv.logger.Error(ctx, "SemesterSettleRecordService syncSemesterSettleRecords panic", nil, zap.Any(pkgLogger.PanicMessage, r))
			}
			if wait != nil {
				wait.Done()
			}
		}()

		srv.syncSemesterSettleRecords(ctx, allRecords, cond)
	}()

	return nil
}

func (srv *SemesterSettleRecordService) syncSemesterSettleRecords(ctx context.Context, allRecords []*bo.SemesterSettleRecord, cond *bo.SyncSemesterSettleRecordCond) {
	wg := &sync.WaitGroup{}
	currentRecordRefIdMap := map[int]struct{}{}
	for _, semesterSettleRecord := range allRecords {
		wg.Add(1)
		ssr := semesterSettleRecord
		srv.executorPool.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					srv.logger.Error(ctx, "SemesterSettleRecordService syncSemesterSettleRecords syncSemesterSettleRecord panic", nil,
						zap.Any(pkgLogger.PanicMessage, r),
						zap.Any("semesterSettleRecord", *ssr),
					)
				}
				wg.Done()
			}()

			if err := srv.syncSemesterSettleRecord(ctx, ssr); err != nil {
				srv.logger.Error(ctx, "SemesterSettleRecordService syncSemesterSettleRecords syncSemesterSettleRecord", err, zap.Int("record_ref_id", ssr.RecordRefId), zap.String("student_name", ssr.KintoneStudentName))
			}
		})

		currentRecordRefIdMap[ssr.RecordRefId] = struct{}{}
	}

	wg.Wait()

	db := srv.DB.Session()
	isDeleted := false
	poSemesterSettleRecordCond := &po.SemesterSettleRecordCond{
		IsDeleted: &isDeleted,
	}
	if cond != nil {
		if cond.StudentName != nil && cond.ParentPhone != nil {
			boStudent, err := srv.studentCommonSrv.GetStudent(ctx, &bo.StudentCond{StudentName: *cond.StudentName, ParentPhone: *cond.ParentPhone})
			if err != nil {
				if errors.Is(err, errs.DbErr.NoRow) {
					srv.logger.Error(ctx, "SemesterSettleRecordService syncSemesterSettleRecords studentCommonSrv.GetStudent", errs.StudentErr.StudentNotFoundErr,
						zap.String("student_name", *cond.StudentName),
						zap.String("parent_phone", *cond.ParentPhone),
					)
					return
				}

				srv.logger.Error(ctx, "SemesterSettleRecordService syncSemesterSettleRecords studentCommonSrv.GetStudent", err,
					zap.String("student_name", *cond.StudentName),
					zap.String("parent_phone", *cond.ParentPhone),
				)

				return
			}

			poSemesterSettleRecordCond.StudentId = boStudent.StudentId
		}
		if cond.StartTime != nil {
			poSemesterSettleRecordCond.StartTime = *cond.StartTime
		}
		if cond.EndTime != nil {
			poSemesterSettleRecordCond.EndTime = *cond.EndTime
		}
	}
	recordRefIds, err := srv.semesterSettleRecordRepo.GetSemesterSettleRecordRefIds(ctx, db, poSemesterSettleRecordCond)
	if err != nil {
		srv.logger.Error(ctx, "SemesterSettleRecordService syncSemesterSettleRecords GetSemesterSettleRecordRefIds", err)
		return
	}

	for _, recordRefId := range recordRefIds {
		if _, found := currentRecordRefIdMap[recordRefId]; !found {
			if err := srv.DeleteSemesterSettleRecord(ctx, &bo.UpdateSemesterSettleRecordCond{RecordRefId: recordRefId}); err != nil {
				srv.logger.Error(ctx, "SemesterSettleRecordService syncSemesterSettleRecords DeleteSemesterSettleRecord", err, zap.Int("record_ref_id", recordRefId))
			}
		}
	}
}

func (srv *SemesterSettleRecordService) syncSemesterSettleRecord(ctx context.Context, data *bo.SemesterSettleRecord) error {
	db := srv.DB.Session()
	poSemesterSettleRecord := &po.SemesterSettleRecordCond{RecordRefId: data.RecordRefId}
	record, err := srv.semesterSettleRecordRepo.GetRecord(ctx, db, poSemesterSettleRecord)
	if err != nil && !errors.Is(err, errs.DbErr.NoRow) {
		return xerrors.Errorf("recordRepo.GetRecord: %w", err)
	}

	boStudentCond := &bo.StudentCond{StudentName: data.StudentName, ParentPhone: data.ParentPhone}
	if record != nil {
		// Update
		boUpdateSemesterSettleRecordCond := &bo.UpdateSemesterSettleRecordCond{RecordRefId: data.RecordRefId}
		boUpdateSemesterSettleRecordData := &bo.UpdateSemesterSettleRecordData{
			StartTime:   data.StartTime,
			EndTime:     data.EndTime,
			ClearPoints: &data.ClearPoints,
		}
		if err := srv.UpdateSemesterSettleRecord(ctx, boUpdateSemesterSettleRecordCond, boStudentCond, boUpdateSemesterSettleRecordData); err != nil {
			return xerrors.Errorf("UpdateSemesterSettleRecord: %w", err)
		}
		//如果在 kintone 上刪除了照理說不會再拿到相同 id 的資料，這裡是為了避免程式邏輯錯誤導致誤刪除了不該刪的記錄
		//所以 API 如果取得到這筆確實存在的記錄但是 DB 卻標示為刪除時，還是重置該記錄的刪除狀態
		if record.IsDeleted || record.DeletedAt != nil {
			if err := srv.semesterSettleRecordRepo.ResetFromDeleted(ctx, db, (po.SemesterSettleRecord{}).TableName(), func(db *gorm.DB) *gorm.DB {
				db.Where("record_id = ?", record.RecordId)
				return db
			}); err != nil {
				return xerrors.Errorf("ResetFromDeleted: %w", err)
			}
		}
	} else {
		// Create
		if err := srv.AddSemesterSettleRecord(ctx, data, boStudentCond); err != nil {
			return xerrors.Errorf("AddSemesterSettleRecord: %w", err)
		}
	}

	return nil
}

func (srv *SemesterSettleRecordService) SettleSemesterPoints(ctx context.Context, cond *bo.SettleSemesterPointsCond) (err error) {
	/* 不是特定結算日，不執行 */
	if !srv.checkIsSettleDate(cond.Date) {
		return nil
	}

	/* 取得結算起始、結束時間 */
	r := dateRange{}
	err = srv.getSemesterRanges(cond.Date, &r)
	if err != nil {
		return xerrors.Errorf("getSemesterRanges: %w", err)
	}

	/* 同步 kintone 學生/點數/購課/點名資料 */
	// 必須先同步學生資料，後續才同步其他跟學生資料相關的資料
	studentWg := sync.WaitGroup{}
	err = srv.studentSrv.BatchSyncStudentsAndUsers(ctx, nil, &studentWg)
	if err != nil {
		return xerrors.Errorf("studentSrv.BatchSyncStudentsAndUsers: %w", err)
	}
	studentWg.Wait()

	wg := sync.WaitGroup{}
	err = srv.pointCardSrv.BatchSyncPointCard(ctx, nil, &wg)
	if err != nil {
		return xerrors.Errorf("pointCardSrv.BatchSyncPointCard: %w", err)
	}

	drCond := bo.SyncDepositRecordCond{
		ChargingDateStart: &r.Start,
		ChargingDateEnd:   &r.End,
	}
	err = srv.depositRecordSrv.BatchSyncDepositRecord(ctx, &drCond, &wg)
	if err != nil {
		return xerrors.Errorf("depositRecordSrv.BatchSyncDepositRecord: %w", err)
	}

	rrCond := bo.SyncReduceRecordCond{
		ClassTimeStart: &r.Start,
		ClassTimeEnd:   &r.End,
	}
	err = srv.reduceRecordSrv.BatchSyncReduceRecord(ctx, &rrCond, &wg)
	if err != nil {
		return xerrors.Errorf("reduceRecordSrv.BatchSyncReduceRecord: %w", err)
	}

	wg.Wait()

	/* 結算、清零剩餘點數 */
	db := srv.DB.Session()
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = xerrors.Errorf("panic on error: %v", r)
			return
		}
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	/* 撈取需要結算的學生 */
	settledStudents, err := srv.studentSrv.GetStudentsSettled(ctx, tx)
	if err != nil {
		return xerrors.Errorf("studentSrv.GetStudentsSettled: %w", err)
	}

	/* 撈取學生已結算記錄 */
	semesterSettleRecordReq := &dto.SemesterSettleRecordReq{
		StartTime: r.Start,
		EndTime:   r.End,
		Limit:     request.GetRecordsBatchLimit,
		Offset:    request.GetRecordsBatchOffset,
	}
	allRecords, err := srv.getAllKintoneSemesterSettleRecords(ctx, semesterSettleRecordReq)
	if err != nil {
		return xerrors.Errorf("getAllKintoneSemesterSettleRecords: %w", err)
	}
	semesterSettleRecordsMap := map[string]struct{}{}
	for _, semesterSettleRecord := range allRecords {
		semesterSettleRecordsMap[semesterSettleRecord.KintoneStudentName] = struct{}{}
	}

	/* 過濾掉已經結算過的學生 */
	studentIds := make([]int64, 0, len(settledStudents))
	filteredSettledStudents := make([]*bo.Student, 0, len(settledStudents))
	for _, student := range settledStudents {
		if _, ok := semesterSettleRecordsMap[strUtil.GetFullStudentName(student.StudentName, student.ParentPhone)]; ok {
			srv.logger.Info(ctx, "SemesterSettleRecordService SettleSemesterPoints: skip settled student", zap.Int64("student_id", student.StudentId), zap.String("student_name", student.StudentName))
			continue
		}
		studentIds = append(studentIds, student.StudentId)
		filteredSettledStudents = append(filteredSettledStudents, student)
	}

	/* 所有學生都已經有結算過(有該學期的結算記錄) */
	if len(studentIds) == 0 {
		srv.logger.Info(ctx, "SemesterSettleRecordService SettleSemesterPoints: all students were settled, checked by semester settle records", zap.Time("startTime", r.Start), zap.Time("endTime", r.End))
		return nil
	}

	/* 撈取未結算學生的過去購買記錄/點名記錄總額, 組出結算資訊 */
	boStudentTotalDepositPointsCond := &bo.StudentTotalDepositPointsCond{
		StudentIds:        studentIds,
		ChargingDateStart: r.Start,
		ChargingDateEnd:   r.End,
	}
	studentTotalDepositPoints, err := srv.depositRecordSrv.GetStudentTotalDepositPoints(ctx, tx, boStudentTotalDepositPointsCond)
	if err != nil {
		return xerrors.Errorf("depositRecordSrv.GetStudentTotalDepositPoints: %w", err)
	}

	boStudentTotalReducePointsCond := &bo.StudentTotalReducePointsCond{
		StudentIds:     studentIds,
		ClassTimeStart: r.Start,
		ClassTimeEnd:   r.End,
	}
	studentTotalReducePoints, err := srv.reduceRecordSrv.GetStudentTotalReducePoints(ctx, tx, boStudentTotalReducePointsCond)
	if err != nil {
		return xerrors.Errorf("reduceRecordSrv.GetStudentTotalReducePoints: %w", err)
	}

	/* 統計並組出結算清單 */
	settlementList := make([]settlementItem, 0, len(studentIds))
	for _, student := range filteredSettledStudents {
		var total float64
		if v, ok := studentTotalDepositPoints[student.StudentId]; ok {
			total += float64(v.TotalDepositPoints)
		}
		if v, ok := studentTotalReducePoints[student.StudentId]; ok {
			total -= v.TotalReducePoints
		}
		if total < 0.0 {
			srv.logger.Warn(ctx, "SemesterSettleRecordService SettleSemesterPoints warning: negative total points",
				zap.String("student_name", student.StudentName),
				zap.Float64("total_deposit_points", float64(studentTotalDepositPoints[student.StudentId].TotalDepositPoints)),
				zap.Float64("total_reduce_points", studentTotalReducePoints[student.StudentId].TotalReducePoints),
			)
			total = 0.0
		}
		settlementList = append(settlementList, settlementItem{
			student:     student,
			clearPoints: total,
		})
	}

	/* 批次新增結算記錄 */
	if err = srv.InsertKintoneSemesterSettleRecords(ctx, settlementList, r); err != nil {
		return xerrors.Errorf("kintoneSemesterSettleRecordRepo.InsertKintoneSemesterSettleRecords: %w", err)
	}

	studentPointCardMap, err := srv.pointCardSrv.GetPointCards(ctx, tx, &bo.GetPointCardCond{StudentIds: studentIds})
	if err != nil {
		return xerrors.Errorf("pointCardSrv.GetPointCards: %w", err)
	}

	boSyncSettledStudentPointCards := make([]*bo.SyncSettledStudentPointCardData, 0, len(settlementList))
	for _, item := range settlementList {
		boSyncSettledStudentPointCards = append(boSyncSettledStudentPointCards, &bo.SyncSettledStudentPointCardData{
			StudentId:          item.student.StudentId,
			KintoneStudentName: strUtil.GetFullStudentName(item.student.StudentName, item.student.ParentPhone),
			ClearPoints:        item.clearPoints,
			RestPoints:         studentPointCardMap[item.student.StudentId].RestPoints - item.clearPoints,
		})
	}

	err = srv.pointCardSrv.SyncSettledStudentPointCards(ctx, tx, boSyncSettledStudentPointCards)
	if err != nil {
		return xerrors.Errorf("pointCardSrv.SyncSettledStudentPointCards: %w", err)
	}

	if err = tx.Commit().Error; err != nil {
		return xerrors.Errorf("Commit: %w", err)
	}

	/* 再同步一次新增的結算記錄 */
	err = srv.BatchSyncSemesterSettleRecord(ctx, &bo.SyncSemesterSettleRecordCond{StartTime: &r.Start, EndTime: &r.End})
	if err != nil {
		srv.logger.Error(ctx, "SemesterSettleRecordService SettleSemesterPoints BatchSyncSemesterSettleRecord failed", err, zap.Time("startTime", r.Start), zap.Time("endTime", r.End))
	}

	return nil
}

func (srv *SemesterSettleRecordService) checkIsSettleDate(t time.Time) bool {
	date := t.Format("01/02")
	return slices.Contains(semesterStartAtDates, date)
}

func (srv *SemesterSettleRecordService) getSemesterRanges(t time.Time, r *dateRange) error {
	nowY := t.Year()
	date := t.Format("01/02")

	idx := slices.Index(semesterStartAtDates, date)
	end := semesterStartAtDates[idx]
	var (
		atStart bool
		start   string
	)
	if idx > 0 {
		start = semesterStartAtDates[idx-1]
	} else {
		start = semesterStartAtDates[len(semesterStartAtDates)-1]
		atStart = true
	}

	t, err := time.ParseInLocation("01/02", end, utils.GetLocation())
	if err != nil {
		return err
	}
	t = t.AddDate(nowY, 0, 0)

	t2, err := time.ParseInLocation("01/02", start, utils.GetLocation())
	if err != nil {
		return err
	}
	if atStart {
		t2 = t2.AddDate(nowY-1, 0, 0)
	} else {
		t2 = t2.AddDate(nowY, 0, 0)
	}

	(*r).Start = t2
	(*r).End = t.Add(-time.Second)

	return nil
}

func (srv *SemesterSettleRecordService) getAllKintoneSemesterSettleRecords(ctx context.Context, req *dto.SemesterSettleRecordReq) ([]*bo.SemesterSettleRecord, error) {
	kintoneSemesterSettleRecords, total, err := srv.getKintoneSemesterSettleRecords(ctx, req)
	if err != nil {
		return nil, xerrors.Errorf("getKintoneSemesterSettleRecords: %w", err)
	}

	allRecords := make([]*bo.SemesterSettleRecord, 0, total)
	allRecords = append(allRecords, kintoneSemesterSettleRecords...)

	offset, limit := req.Offset, req.Limit
	for ((offset * limit) + limit) < total {
		offset += 1
		req.Offset = offset * limit
		kintoneSemesterSettleRecords, _, err = srv.getKintoneSemesterSettleRecords(ctx, req)
		if err != nil {
			return nil, xerrors.Errorf("getKintoneSemesterSettleRecords: %w", err)
		}
		allRecords = append(allRecords, kintoneSemesterSettleRecords...)
	}

	return allRecords, nil
}

func (srv *SemesterSettleRecordService) getKintoneSemesterSettleRecords(ctx context.Context, cond *dto.SemesterSettleRecordReq) ([]*bo.SemesterSettleRecord, int, error) {
	poSemesterSettleRecordRes, err := srv.kintoneSemesterSettleRecordRepo.GetKintoneSemesterSettleRecords(ctx, cond)
	if err != nil {
		return nil, 0, xerrors.Errorf("kintoneSemesterSettleRecordRepo.GetKintoneSemesterSettleRecords: %w", err)
	}

	total, err := strconv.ParseInt(poSemesterSettleRecordRes.TotalCount, 10, 64)
	if err != nil {
		return nil, 0, xerrors.Errorf("strconv.ParseInt: %w", err)
	}

	boSemesterSettleRecords := make([]*bo.SemesterSettleRecord, 0, len(poSemesterSettleRecordRes.Records))

	for _, record := range poSemesterSettleRecordRes.Records {
		boSemesterSettleRecord, err := record.ToSemesterSettleRecord()
		if err != nil {
			return nil, 0, xerrors.Errorf("record.ToSemesterSettleRecord: %w", err)
		}
		boSemesterSettleRecords = append(boSemesterSettleRecords, boSemesterSettleRecord)
	}

	return boSemesterSettleRecords, int(total), nil
}

func (srv *SemesterSettleRecordService) InsertKintoneSemesterSettleRecords(ctx context.Context, settlementList []settlementItem, r dateRange) error {
	batchRecords := make([]dto.InsertSemesterSettleRecord, 0, kintone.BatchInsertRecordsMaxLimit)
	req := &dto.InsertSemesterSettleRecordsReq{}
	err := utils.RunInBatch(len(settlementList), kintone.BatchInsertRecordsMaxLimit, func(batchIndex int, start int, end int) error {
		for _, item := range settlementList[start:end] {
			insertRecord := dto.InsertSemesterSettleRecord{}
			insertRecord.StudentName.Value = strUtil.GetFullStudentName(item.student.StudentName, item.student.ParentPhone)
			// kintone 應用起始時間/結束時間只記錄日期
			insertRecord.StartTime.Value = r.Start.Format(time.DateOnly)
			insertRecord.EndTime.Value = r.End.Format(time.DateOnly)
			insertRecord.ClearPoints.Value = strconv.FormatFloat(item.clearPoints, 'f', -1, 64)
			batchRecords = append(batchRecords, insertRecord)
		}

		req.Records = batchRecords
		if _, err := srv.kintoneSemesterSettleRecordRepo.InsertKintoneSemesterSettleRecords(ctx, req); err != nil {
			return xerrors.Errorf("kintoneSemesterSettleRecordRepo.InsertKintoneSemesterSettleRecords batch [%d]: %w", batchIndex, err)
		}

		batchRecords = batchRecords[:0]
		srv.logger.Info(ctx, fmt.Sprintf("InsertKintoneSemesterSettleRecords: batch %d succeeded", batchIndex))

		time.Sleep(100 * time.Millisecond) // 避免短時間發送太多請求造成 API 失敗

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
