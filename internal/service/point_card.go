package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/panjf2000/ants"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
	"jaystar/internal/constant/request"
	"jaystar/internal/database"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/model/po"
	"jaystar/internal/utils/errs"
	"jaystar/internal/utils/pool"
	"jaystar/internal/utils/strUtil"
	"strconv"
	"sync"
	"time"
)

func ProvidePointCardService(
	kintonePointCardRepo interfaces.IKintonePointCardRepo,
	pointCardRepo interfaces.IPointCardRepo,
	studentRepo interfaces.IStudentRepo,
	DB database.IPostgresDB,
	logger logger.ILogger,
) *PointCardService {
	return &PointCardService{
		kintonePointCardRepo: kintonePointCardRepo,
		pointCardRepo:        pointCardRepo,
		studentRepo:          studentRepo,
		DB:                   DB,
		logger:               logger,
		executorPool:         pool.NewExecutorPool(30),
	}
}

type PointCardService struct {
	kintonePointCardRepo interfaces.IKintonePointCardRepo
	pointCardRepo        interfaces.IPointCardRepo
	studentRepo          interfaces.IStudentRepo
	DB                   database.IPostgresDB
	logger               logger.ILogger
	executorPool         *ants.Pool `wire:"-"`
}

func (srv *PointCardService) UpdateKintonePointCardStudentName(ctx context.Context, oldPointCardName string, newPointCardName string) error {
	getReq := &dto.GetPointCardReq{StudentName: oldPointCardName, Limit: 1, Offset: 0}
	pointCardRes, err := srv.kintonePointCardRepo.GetPointCards(ctx, getReq)
	if err != nil {
		return xerrors.Errorf("kintonePointCardRepo.GetPointCards: %w", err)
	}
	if len(pointCardRes.Records) == 0 {
		return fmt.Errorf("cannot find point card record for student: %s", oldPointCardName)
	}

	id, err := pointCardRes.Records[0].Id.ToId()
	if err != nil {
		return xerrors.Errorf("pointCardRes.Records[0].Id.ToId student: %s, id: %s, err: %w", oldPointCardName, pointCardRes.Records[0].Id.Value, err)
	}

	updateReq := &dto.UpdatePointCardReq{}
	updateReq.Id = id
	updateReq.Record.StudentName.Value = newPointCardName
	_, err = srv.kintonePointCardRepo.UpdatePointCard(ctx, updateReq)
	if err != nil {
		return xerrors.Errorf("kintonePointCardRepo.UpdatePointCard: %w", err)
	}

	return nil
}

func (srv *PointCardService) GetPointCards(ctx context.Context, db *gorm.DB, cond *bo.GetPointCardCond) (map[int64]*bo.PointCard, error) {
	poPointCardCond := &po.PointCardCond{
		StudentIds: cond.StudentIds,
	}

	if db == nil {
		db = srv.DB.Session()
	}

	poPointCards, err := srv.pointCardRepo.GetPointCards(ctx, db, poPointCardCond)
	if err != nil {
		return nil, xerrors.Errorf("pointCardRepo.GetPointCards: %w", err)
	}

	boPointCards := make(map[int64]*bo.PointCard)
	for _, card := range poPointCards {
		boPointCards[card.StudentId] = &bo.PointCard{
			RecordId:    card.RecordId,
			RecordRefId: card.RecordRefId,
			StudentId:   card.StudentId,
			RestPoints:  card.RestPoints,
		}
	}

	return boPointCards, nil
}

func (srv *PointCardService) AddPointCard(ctx context.Context, data *bo.PointCard, studentCond *bo.StudentCond) error {
	db := srv.DB.Session()

	student, err := srv.studentRepo.GetStudent(ctx, db, &po.StudentCond{StudentName: studentCond.StudentName, ParentPhone: studentCond.ParentPhone})
	if err != nil {
		if errors.Is(err, errs.DbErr.NoRow) {
			return xerrors.Errorf("studentRepo.GetStudent: %w", errs.StudentErr.StudentNotFoundErr)
		}

		return xerrors.Errorf("studentRepo.GetStudent: %w", err)
	}

	poPointCardRecord := &po.PointCard{
		RecordRefId: data.RecordRefId,
		StudentId:   student.StudentId,
		RestPoints:  data.RestPoints,
	}

	if err := srv.pointCardRepo.AddPointCard(ctx, db, poPointCardRecord); err != nil {
		return xerrors.Errorf("pointCardRepo.AddPointCard: %w", err)
	}

	return nil
}

func (srv *PointCardService) UpdatePointCard(ctx context.Context, cond *bo.UpdatePointCardCond, studentCond *bo.StudentCond, data *bo.UpdatePointCardRecordData) error {
	db := srv.DB.Session()

	student, err := srv.studentRepo.GetStudent(ctx, db, &po.StudentCond{StudentName: studentCond.StudentName, ParentPhone: studentCond.ParentPhone})
	if err != nil {
		if errors.Is(err, errs.DbErr.NoRow) {
			return xerrors.Errorf("studentRepo.GetStudent: %w", errs.StudentErr.StudentNotFoundErr)
		}

		return xerrors.Errorf("studentRepo.GetStudent: %w", err)
	}

	updatePointCardCond := &po.UpdatePointCardCond{RecordRefId: cond.RecordRefId}
	poUpdatePointCardData := &po.UpdatePointCardData{
		StudentId:  student.StudentId,
		RestPoints: data.RestPoints,
	}
	if err := srv.pointCardRepo.UpdatePointCard(ctx, db, updatePointCardCond, poUpdatePointCardData); err != nil {
		return xerrors.Errorf("pointCardRepo.UpdatePointCard: %w", err)
	}

	return nil
}

func (srv *PointCardService) DeletePointCard(ctx context.Context, cond *bo.UpdatePointCardCond) error {
	deleted := true
	poUpdatePointCardCond := &po.UpdatePointCardCond{
		RecordRefId: cond.RecordRefId,
	}
	now := time.Now()
	poUpdatePointCardData := &po.UpdatePointCardData{
		IsDeleted: &deleted,
		DeletedAt: &now,
	}

	db := srv.DB.Session()
	if err := srv.pointCardRepo.UpdatePointCard(ctx, db, poUpdatePointCardCond, poUpdatePointCardData); err != nil {
		return xerrors.Errorf("pointCardRepo.UpdatePointCard: %w", err)
	}

	return nil
}

func (srv *PointCardService) BatchSyncPointCard(ctx context.Context, cond *bo.SyncPointCardCond, wait ...*sync.WaitGroup) error {
	getPointCardReq := &dto.GetPointCardReq{
		Limit:  request.GetRecordsBatchLimit,
		Offset: request.GetRecordsBatchOffset,
	}

	var (
		total        int
		err          error
		boPointCards []*bo.PointCard
	)

	if cond != nil {
		if cond.StudentName != nil && cond.ParentPhone != nil {
			getPointCardReq.StudentName = strUtil.GetFullStudentName(*cond.StudentName, *cond.ParentPhone)
		}
		if cond.Limit != nil {
			getPointCardReq.Limit = *cond.Limit
		}
		if cond.Offset != nil {
			getPointCardReq.Offset = *cond.Offset
		}
	}

	boPointCards, total, err = srv.getKintonePointCards(ctx, getPointCardReq)
	if err != nil {
		return xerrors.Errorf("getKintonePointCards: %w", err)
	}

	allRecords := make([]*bo.PointCard, 0, total)
	allRecords = append(allRecords, boPointCards...)

	offset, limit := getPointCardReq.Offset, getPointCardReq.Limit
	for ((offset * limit) + limit) < total {
		offset += 1
		getPointCardReq.Offset = offset * limit
		boPointCards, _, err = srv.getKintonePointCards(ctx, getPointCardReq)
		if err != nil {
			return xerrors.Errorf("getKintonePointCards: %w", err)
		}
		allRecords = append(allRecords, boPointCards...)
	}

	var wg *sync.WaitGroup
	if len(wait) > 0 {
		wg = wait[0]
		wg.Add(1)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				srv.logger.Error(ctx, "pointCardService syncPointCards panic", nil, zap.Any(logger.PanicMessage, r))
			}
			if wg != nil {
				wg.Done()
			}
		}()
		srv.syncPointCards(ctx, allRecords, cond)
	}()

	return nil
}

func (srv *PointCardService) SyncSettledStudentPointCards(ctx context.Context, db *gorm.DB, data []*bo.SyncSettledStudentPointCardData) error {
	if len(data) == 0 {
		return nil
	}

	kintoneUpdatePointCardsRecords := make([]dto.UpdatePointCardsRecord, 0, len(data))
	for _, updatePointCardsData := range data {
		updatePointCardCond := &po.UpdatePointCardCond{StudentId: updatePointCardsData.StudentId}
		poUpdatePointCardData := &po.UpdatePointCardData{
			RestPoints: &updatePointCardsData.RestPoints,
		}
		if err := srv.pointCardRepo.UpdatePointCard(ctx, db, updatePointCardCond, poUpdatePointCardData); err != nil {
			return xerrors.Errorf("pointCardRepo.UpdatePointCard: %w", err)
		}

		kintoneUpdatePointCardsRecord := dto.UpdatePointCardsRecord{
			UpdateKey: struct {
				Field string `json:"field"`
				Value string `json:"value"`
			}{Field: "studentName", Value: updatePointCardsData.KintoneStudentName},
		}

		kintoneUpdatePointCardRecord := dto.UpdatePointCardsRecordValue{}
		kintoneUpdatePointCardRecord.ClearPoints.Value = strconv.FormatFloat(updatePointCardsData.ClearPoints, 'f', -1, 64)

		kintoneUpdatePointCardsRecord.Record = kintoneUpdatePointCardRecord
		kintoneUpdatePointCardsRecords = append(kintoneUpdatePointCardsRecords, kintoneUpdatePointCardsRecord)
	}

	req := &dto.UpdatePointCardsReq{
		Records: kintoneUpdatePointCardsRecords,
	}
	_, err := srv.kintonePointCardRepo.UpdatePointCards(ctx, req)
	if err != nil {
		return xerrors.Errorf("kintonePointCardRepo.UpdatePointCards: %w", err)
	}

	return nil
}

func (srv *PointCardService) getKintonePointCards(ctx context.Context, getPointCardReq *dto.GetPointCardReq) ([]*bo.PointCard, int, error) {
	kintonePointCardRes, err := srv.kintonePointCardRepo.GetPointCards(ctx, getPointCardReq)
	if err != nil {
		return nil, 0, xerrors.Errorf("kintonePointCardRepo.GetPointCards: %w", err)
	}
	total, err := strconv.Atoi(kintonePointCardRes.TotalCount)
	if err != nil {
		return nil, 0, xerrors.Errorf("strconv.ParseInt: %w", err)
	}

	boPointCardRecords := make([]*bo.PointCard, 0, len(kintonePointCardRes.Records))

	for _, record := range kintonePointCardRes.Records {
		boPointCard, err := record.ToPointCard()
		if err != nil {
			return nil, 0, xerrors.Errorf("record.ToPointCard: %w", err)
		}
		boPointCardRecords = append(boPointCardRecords, boPointCard)
	}

	return boPointCardRecords, total, nil
}

func (srv *PointCardService) syncPointCards(ctx context.Context, allRecords []*bo.PointCard, cond *bo.SyncPointCardCond) {
	wg := &sync.WaitGroup{}
	currentRecordRefIdMap := map[int]struct{}{}
	for _, pointCard := range allRecords {
		wg.Add(1)
		p := pointCard
		srv.executorPool.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					srv.logger.Error(ctx, "PointCardService syncPointCards syncPointCard panic", nil,
						zap.Any(logger.PanicMessage, r),
						zap.Any("pointCard", *p),
					)
				}
				wg.Done()
			}()

			if err := srv.syncPointCard(ctx, p); err != nil {
				srv.logger.Error(ctx, "PointCardService syncPointCards syncPointCard", err, zap.Int("record_ref_id", p.RecordRefId), zap.String("student_name", p.KintoneStudentName))
			}
		})

		currentRecordRefIdMap[p.RecordRefId] = struct{}{}
	}

	wg.Wait()

	db := srv.DB.Session()
	isDeleted := false
	poPointCardCond := &po.PointCardCond{
		IsDeleted: &isDeleted,
	}
	if cond != nil {
		if cond.StudentName != nil && cond.ParentPhone != nil {
			student, err := srv.studentRepo.GetStudent(ctx, db, &po.StudentCond{StudentName: *cond.StudentName, ParentPhone: *cond.StudentName})
			if err != nil {
				if errors.Is(err, errs.DbErr.NoRow) {
					srv.logger.Error(ctx, "PointCardService syncPointCards studentRepo.GetStudent", errs.StudentErr.StudentNotFoundErr, zap.String("student_name", *cond.StudentName), zap.String("parent_phone", *cond.ParentPhone))
					return
				}

				srv.logger.Error(ctx, "PointCardService syncPointCards studentRepo.GetStudent", err)
				return
			}

			poPointCardCond.StudentIds = []int64{student.StudentId}
		}
	}

	recordRefIds, err := srv.pointCardRepo.GetPointCardRefIds(ctx, db, poPointCardCond)
	if err != nil {
		srv.logger.Error(ctx, "PointCardService syncPointCards pointCardRepo.GetPointCardRefIds", err)
		return
	}

	for _, recordRefId := range recordRefIds {
		if _, found := currentRecordRefIdMap[recordRefId]; !found {
			if err := srv.DeletePointCard(ctx, &bo.UpdatePointCardCond{RecordRefId: recordRefId}); err != nil {
				srv.logger.Error(ctx, "PointCardService syncPointCards DeletePointCard", err, zap.Int("record_ref_id", recordRefId))
			}
		}
	}
}

func (srv *PointCardService) syncPointCard(ctx context.Context, data *bo.PointCard) error {
	db := srv.DB.Session()
	poPointCardCond := &po.PointCardCond{RecordRefId: data.RecordRefId}
	record, err := srv.pointCardRepo.GetPointCard(ctx, db, poPointCardCond)
	if err != nil && !errors.Is(err, errs.DbErr.NoRow) {
		return xerrors.Errorf("pointCardRepo.GetPointCard: %w", err)
	}

	boStudentCond := &bo.StudentCond{StudentName: data.StudentName, ParentPhone: data.ParentPhone}
	if record != nil {
		// Update
		boUpdatePointCardCond := &bo.UpdatePointCardCond{RecordRefId: data.RecordRefId}
		boUpdatePointCardRecordData := &bo.UpdatePointCardRecordData{
			RestPoints: &data.RestPoints,
		}
		if err := srv.UpdatePointCard(ctx, boUpdatePointCardCond, boStudentCond, boUpdatePointCardRecordData); err != nil {
			return xerrors.Errorf("UpdatePointCard: %w", err)
		}
		//如果在 kintone 上刪除了照理說不會再拿到相同 id 的資料，這裡是為了避免程式邏輯錯誤導致誤刪除了不該刪的記錄
		//所以 API 如果取得到這筆確實存在的記錄但是 DB 卻標示為刪除時，還是重置該記錄的刪除狀態
		if record.IsDeleted || record.DeletedAt != nil {
			if err := srv.pointCardRepo.ResetFromDeleted(ctx, db, (po.PointCard{}).TableName(), func(db *gorm.DB) *gorm.DB {
				db.Where("record_id = ?", record.RecordId)
				return db
			}); err != nil {
				return xerrors.Errorf("ResetFromDeleted: %w", err)
			}
		}
	} else {
		// Create
		if err := srv.AddPointCard(ctx, data, boStudentCond); err != nil {
			return xerrors.Errorf("AddPointCard: %w", err)
		}
	}

	return nil
}
