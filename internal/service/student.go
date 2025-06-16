package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/SeanZhenggg/go-utils/snowflake/autoId"
	"github.com/forgoer/openssl"
	"github.com/panjf2000/ants"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"gorm.io/gorm"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/constant/request"
	"jaystar/internal/constant/user"
	"jaystar/internal/database"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/model/po"
	"jaystar/internal/utils/errs"
	"jaystar/internal/utils/pool"
	"jaystar/internal/utils/strUtil"
	"reflect"
	"strings"
	"sync"
	"time"
)

func ProvideStudentService(
	db database.IPostgresDB,
	studentRepo interfaces.IStudentRepo,
	userCommonSrv interfaces.IUserCommonSrv,
	studentCommonSrv interfaces.IStudentCommonSrv,
	logger logger.ILogger,
	pointCardCommonSrv interfaces.IPointCardSrv,
	depositCommonSrv interfaces.IDepositRecordCommonSrv,
	reduceCommonSrv interfaces.IReduceRecordCommonSrv,
	scheduleCommonSrv interfaces.IScheduleCommonSrv,
) *StudentService {
	return &StudentService{
		DB:                 db,
		studentRepo:        studentRepo,
		userCommonSrv:      userCommonSrv,
		studentCommonSrv:   studentCommonSrv,
		logger:             logger,
		pointCardCommonSrv: pointCardCommonSrv,
		depositCommonSrv:   depositCommonSrv,
		reduceCommonSrv:    reduceCommonSrv,
		scheduleCommonSrv:  scheduleCommonSrv,
		executorPool:       pool.NewExecutorPool(30),
	}
}

type StudentService struct {
	studentRepo          interfaces.IStudentRepo
	DB                   database.IPostgresDB
	userCommonSrv        interfaces.IUserCommonSrv
	studentCommonSrv     interfaces.IStudentCommonSrv
	pointCardCommonSrv   interfaces.IPointCardSrv
	depositCommonSrv     interfaces.IDepositRecordCommonSrv
	reduceCommonSrv      interfaces.IReduceRecordCommonSrv
	scheduleCommonSrv    interfaces.IScheduleCommonSrv
	logger               logger.ILogger
	kintonePointCardRepo interfaces.IKintonePointCardRepo
	executorPool         *ants.Pool `wire:"-"`
}

func (srv *StudentService) GetStudents(ctx context.Context, cond *bo.StudentCond) ([]*bo.Student, *po.PagerResult, error) {
	if reflect.ValueOf(cond).Elem().IsZero() {
		return nil, nil, errs.DbErr.NoRow
	}

	poStudentCond := &po.StudentCond{}

	if cond.UserId != 0 {
		poStudentCond.UserId = cond.UserId
	}

	// 1. 有帶 student_id
	// 2. 有帶 student_ref_id
	// 3. 有帶 student_name 或 parent_phone
	if cond.StudentId != 0 {
		poStudentCond.StudentId = cond.StudentId
	} else if cond.StudentRefId != 0 {
		poStudentCond.StudentRefId = cond.StudentRefId
	} else {
		if cond.StudentName != "" {
			poStudentCond.StudentName = cond.StudentName
		}
		if cond.ParentPhone != "" {
			poStudentCond.ParentPhone = cond.ParentPhone
		}
	}
	if cond.IsDeleted != nil {
		poStudentCond.IsDeleted = cond.IsDeleted
	} else {
		isDeleted := false
		poStudentCond.IsDeleted = &isDeleted
	}
	if cond.Mode != nil {
		poStudentCond.Mode = cond.Mode.ToKey()
	}
	if cond.IsSettleNormally != nil {
		poStudentCond.IsSettleNormally = cond.IsSettleNormally
	}

	// 如果有帶分頁資訊，by 分頁查詢
	db := srv.DB.Session()
	poStudents, err := srv.studentRepo.GetStudents(ctx, db, poStudentCond, cond.Pager)
	if err != nil {
		return nil, nil, xerrors.Errorf("studentRepo.GetStudents: %w", err)
	}

	pagerResult := &po.PagerResult{}
	if cond.Pager != nil {
		pagerResult, err = srv.studentRepo.GetStudentsPager(ctx, db, poStudentCond, cond.Pager)
		if err != nil {
			return nil, nil, xerrors.Errorf("studentRepo.GetStudents: %w", err)
		}
	}

	boStudents := make([]*bo.Student, 0, len(poStudents))
	for _, poStudent := range poStudents {
		boStudents = append(boStudents, &bo.Student{
			StudentId:        poStudent.StudentId,
			StudentRefId:     poStudent.StudentRefId,
			UserId:           poStudent.UserId,
			StudentName:      poStudent.StudentName,
			ParentName:       poStudent.ParentName,
			ParentPhone:      poStudent.ParentPhone,
			Balance:          poStudent.Balance,
			Mode:             kintone.ModeToEnum(poStudent.Mode),
			IsSettleNormally: poStudent.IsSettleNormally,
			IsDeleted:        poStudent.IsDeleted,
			CreatedAt:        poStudent.CreatedAt,
			UpdatedAt:        poStudent.UpdatedAt,
			DeletedAt:        poStudent.DeletedAt,
		})
	}

	return boStudents, pagerResult, nil
}

func (srv *StudentService) GetStudentsSettled(ctx context.Context, db *gorm.DB) ([]*bo.Student, error) {
	isDeleted := false
	isSettledNormally := true
	poStudents, err := srv.studentRepo.GetStudents(ctx, db, &po.StudentCond{
		Mode:             kintone.ModeSemester.ToKey(),
		IsDeleted:        &isDeleted,
		IsSettleNormally: &isSettledNormally,
	}, nil)
	if err != nil {
		return nil, xerrors.Errorf("studentRepo.GetStudentsSettled: %w", err)
	}

	boStudents := make([]*bo.Student, 0, len(poStudents))
	for _, poStudent := range poStudents {
		boStudents = append(boStudents, &bo.Student{
			StudentId:        poStudent.StudentId,
			StudentRefId:     poStudent.StudentRefId,
			UserId:           poStudent.UserId,
			StudentName:      poStudent.StudentName,
			ParentName:       poStudent.ParentName,
			ParentPhone:      poStudent.ParentPhone,
			IsDeleted:        poStudent.IsDeleted,
			Balance:          poStudent.Balance,
			Mode:             kintone.ModeToEnum(poStudent.Mode),
			IsSettleNormally: poStudent.IsSettleNormally,
			CreatedAt:        poStudent.CreatedAt,
			UpdatedAt:        poStudent.UpdatedAt,
		})
	}

	return boStudents, nil
}

func (srv *StudentService) UserRegisterAndCreateStudent(ctx context.Context, data *bo.Student) error {
	boDefaultCreateUserData, err := srv.genDefaultCreateUserData(ctx, data)
	if err != nil {
		return xerrors.Errorf("genDefaultCreateUserData: %w", err)
	}

	if err := srv.userCommonSrv.CreateUserAndStudent(ctx, boDefaultCreateUserData, data); err != nil {
		return xerrors.Errorf("userCommonSrv.CreateUserAndStudent: %w", err)
	}

	return nil
}

func (srv *StudentService) addStudent(ctx context.Context, userId int64, data *bo.Student) error {
	studentId, err := autoId.DefaultSnowFlake.GenNextId()
	if err != nil {
		return xerrors.Errorf("autoId.DefaultSnowFlake.GenNextId: %w", err)
	}

	poStudent := &po.Student{
		StudentId:        studentId,
		StudentRefId:     data.StudentRefId,
		UserId:           userId,
		StudentName:      data.StudentName,
		ParentName:       data.ParentName,
		ParentPhone:      data.ParentPhone,
		Mode:             data.Mode.ToKey(),
		IsSettleNormally: data.IsSettleNormally,
	}

	db := srv.DB.Session()
	if err := srv.studentRepo.CreateStudent(ctx, db, poStudent); err != nil {
		return xerrors.Errorf("studentRepo.CreateStudent: %w", err)
	}

	return nil
}

func (srv *StudentService) UpdateStudent(ctx context.Context, data *bo.Student) error {
	// 檢查是否有異動學生姓名 or 家長電話
	dbStudent, err := srv.studentCommonSrv.GetStudent(ctx, &bo.StudentCond{StudentRefId: data.StudentRefId})
	if err != nil {
		return xerrors.Errorf("studentCommonSrv.GetStudent: %w", err)
	}

	// 有，執行其他相關記錄資料的更新，修改以下幾個應用程式的對應資料
	if dbStudent.StudentName != data.StudentName || dbStudent.ParentPhone != data.ParentPhone {
		// 點數管理
		oldPointCardName := strUtil.GetFullStudentName(dbStudent.StudentName, dbStudent.ParentPhone)
		newPointCardName := strUtil.GetFullStudentName(data.StudentName, data.ParentPhone)
		err = srv.pointCardCommonSrv.UpdateKintonePointCardStudentName(ctx, oldPointCardName, newPointCardName)
		if err != nil {
			return xerrors.Errorf("pointCardCommonSrv.UpdateKintonePointCardStudentName: %w", err)
		}
		// 購課記錄
		err = srv.depositCommonSrv.UpdateKintoneDepositRecordsStudentNames(ctx, oldPointCardName, newPointCardName)
		if err != nil {
			if !errors.Is(err, errs.KintoneErr.ResponseEmptyError) {
				return xerrors.Errorf("depositCommonSrv.UpdateKintoneDepositRecordsStudentNames: %w", err)
			} else {
				srv.logger.Warn(ctx, "studentService UpdateStudent depositCommonSrv.UpdateKintoneDepositRecordsStudentNames", zap.Error(err), zap.String("oldPointCardName", oldPointCardName), zap.String("newPointCardName", newPointCardName))
			}
		}
		// 點名管理
		err = srv.reduceCommonSrv.UpdateKintoneReduceRecordsStudentNames(ctx, oldPointCardName, newPointCardName)
		if err != nil {
			if !errors.Is(err, errs.KintoneErr.ResponseEmptyError) {
				return xerrors.Errorf("reduceCommonSrv.UpdateKintoneReduceRecordsStudentNames: %w", err)
			} else {
				srv.logger.Warn(ctx, "studentService UpdateStudent reduceCommonSrv.UpdateKintoneReduceRecordsStudentNames", zap.Error(err), zap.String("oldPointCardName", oldPointCardName), zap.String("newPointCardName", newPointCardName))
			}
		}
		// 課表管理合併點名
		err = srv.scheduleCommonSrv.UpdateKintoneSchedulesStudentNames(ctx, oldPointCardName, newPointCardName)
		if err != nil {
			if !errors.Is(err, errs.KintoneErr.ResponseEmptyError) {
				return xerrors.Errorf("scheduleCommonSrv.UpdateKintoneSchedulesStudentNames: %w", err)
			} else {
				srv.logger.Warn(ctx, "studentService UpdateStudent scheduleCommonSrv.UpdateKintoneSchedulesStudentNames", zap.Error(err), zap.String("oldPointCardName", oldPointCardName), zap.String("newPointCardName", newPointCardName))
			}
		}
	}

	// 綁定家長帳號
	userId, err := srv.getOrCreateUserRetId(ctx, data)
	if err != nil {
		return xerrors.Errorf("getOrCreateUserRetId: %w", err)
	}

	// 更新 db 學生資料
	err = srv.updateStudent(ctx,
		&bo.UpdateStudentCond{RecordRefId: data.StudentRefId},
		&bo.UpdateStudentData{
			StudentName:      data.StudentName,
			ParentName:       data.ParentName,
			ParentPhone:      data.ParentPhone,
			UserId:           userId,
			Mode:             &data.Mode,
			IsSettleNormally: &data.IsSettleNormally,
		},
	)
	if err != nil {
		return xerrors.Errorf("updateStudent: %w", err)
	}

	return nil
}

func (srv *StudentService) getOrCreateUserRetId(ctx context.Context, newData *bo.Student) (int64, error) {
	dbUser, err := srv.userCommonSrv.GetUser(ctx, &bo.UserCond{Accounts: []string{newData.ParentPhone}})
	if err != nil && !errors.Is(err, errs.DbErr.NoRow) {
		return 0, xerrors.Errorf("userCommonSrv.GetUser: %w", err)
	}

	var userId int64

	if dbUser != nil {
		userId = dbUser.UserId
	} else {
		userData, err := srv.genDefaultCreateUserData(ctx, newData)
		if err != nil {
			return 0, xerrors.Errorf("genDefaultCreateUserData: %w", err)
		}

		userId, err = srv.userCommonSrv.CreateUser(ctx, userData)
		if err != nil {
			return 0, xerrors.Errorf("userCommonSrv.CreateUser: %w", err)
		}
	}
	return userId, nil
}

func (srv *StudentService) updateStudent(ctx context.Context, cond *bo.UpdateStudentCond, data *bo.UpdateStudentData) error {
	db := srv.DB.Session()
	poUpdateStudentCond := &po.UpdateStudentCond{StudentRefId: cond.RecordRefId}

	poUpdateStudentData := &po.UpdateStudentData{
		UserId:           data.UserId,
		StudentName:      data.StudentName,
		ParentName:       data.ParentName,
		ParentPhone:      data.ParentPhone,
		IsSettleNormally: data.IsSettleNormally,
	}

	if data.Mode != nil {
		poUpdateStudentData.Mode = data.Mode.ToKey()
	}

	if err := srv.studentRepo.UpdateStudent(ctx, db, poUpdateStudentCond, poUpdateStudentData); err != nil {
		return xerrors.Errorf("studentRepo.UpdateStudent: %w", err)
	}

	return nil
}

func (srv *StudentService) DeleteStudent(ctx context.Context, studentRefId int) error {
	err := srv.deleteStudent(ctx, studentRefId)
	if err != nil {
		return xerrors.Errorf("deleteStudent: %w", err)
	}

	updatedUsers, err := srv.userCommonSrv.DeactivateUserWithNoActiveStudent(ctx)
	if err != nil {
		return xerrors.Errorf("userCommonSrv.DeactivateUserWithNoActiveStudent: %w", err)
	}

	if len(updatedUsers) > 0 {
		userId := make([]string, 0, len(updatedUsers))
		for _, updatedUser := range updatedUsers {
			userId = append(userId, fmt.Sprintf("{id: %d, account: %s}", updatedUser.UserId, updatedUser.Account))
		}
		srv.logger.Info(ctx, "studentService DeleteStudent updatedUsers", zap.Strings("userIds", userId))
	}

	return nil
}

func (srv *StudentService) deleteStudent(ctx context.Context, studentRefId int) error {
	if studentRefId == 0 {
		return xerrors.Errorf("invalid student ref id: %w", errs.StudentErr.StudentIdInvalidErr)
	}

	deleted := true
	poUpdateStudentCond := &po.UpdateStudentCond{StudentRefId: studentRefId}
	now := time.Now()
	poUpdateStudentData := &po.UpdateStudentData{
		IsDeleted: &deleted,
		DeletedAt: &now,
	}

	db := srv.DB.Session()
	if err := srv.studentRepo.UpdateStudent(ctx, db, poUpdateStudentCond, poUpdateStudentData); err != nil {
		return xerrors.Errorf("studentRepo.UpdateStudent: %w", err)
	}

	return nil
}

func (srv *StudentService) BatchSyncStudentsAndUsers(ctx context.Context, cond *bo.SyncStudentCond, wg ...*sync.WaitGroup) error {
	boStudentReq := &dto.StudentReq{}

	var (
		total           int
		err             error
		kintoneStudents []*bo.Student
	)
	boStudentReq.Limit = request.GetRecordsBatchLimit
	boStudentReq.Offset = request.GetRecordsBatchOffset

	if cond != nil {
		if cond.StudentName != nil {
			boStudentReq.StudentName = *cond.StudentName
		}
		if cond.ParentPhone != nil {
			boStudentReq.ParentPhone = *cond.ParentPhone
		}
		if cond.Limit != nil {
			boStudentReq.Limit = *cond.Limit
		}
		if cond.Offset != nil {
			boStudentReq.Offset = *cond.Offset
		}
	}

	kintoneStudents, total, err = srv.studentCommonSrv.GetKintoneStudents(ctx, boStudentReq)
	if err != nil {
		return xerrors.Errorf("studentService BatchSyncStudentsAndUsers studentCommonSrv.GetKintoneStudents: %w", err)
	}
	allStudents := make([]*bo.Student, 0, total)
	allStudents = append(allStudents, kintoneStudents...)

	offset, limit := boStudentReq.Offset, boStudentReq.Limit
	for ((offset * limit) + limit) < total {
		offset += 1
		boStudentReq.Offset = offset * limit
		kintoneStudents, _, err = srv.studentCommonSrv.GetKintoneStudents(ctx, boStudentReq)
		if err != nil {
			return xerrors.Errorf("studentService BatchSyncStudentsAndUsers studentCommonSrv.GetKintoneStudents: %w", err)
		}
		allStudents = append(allStudents, kintoneStudents...)
	}

	var wait *sync.WaitGroup
	if len(wg) > 0 {
		wait = wg[0]
		wait.Add(1)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				srv.logger.Error(ctx, "studentService syncStudentsOrUsers panic", nil,
					zap.Any(logger.PanicMessage, r),
				)
			}

			if wait != nil {
				wait.Done()
			}
		}()
		srv.syncStudentsOrUsers(ctx, allStudents, cond)
	}()

	return nil
}

// genDefaultCreateUserData 創建預設使用者資料
// 帳號:手機，密碼:手機
func (srv *StudentService) genDefaultCreateUserData(ctx context.Context, data *bo.Student) (*bo.CreateUserData, error) {
	// 創建預設使用者。帳號:手機，密碼:手機
	encrypted, err := openssl.AesECBEncrypt([]byte(data.ParentPhone), []byte(user.PasswordAesECBKey), openssl.PKCS7_PADDING)
	if err != nil {
		return nil, xerrors.Errorf("openssl.AesECBEncrypt: %w", err)
	}
	encryptedPwd := base64.StdEncoding.EncodeToString(encrypted)

	boCreateUserData := &bo.CreateUserData{
		Account:  data.ParentPhone,
		Password: encryptedPwd,
	}

	return boCreateUserData, nil
}

func (srv *StudentService) syncStudentsOrUsers(ctx context.Context, allRecords []*bo.Student, cond *bo.SyncStudentCond) {
	wg := &sync.WaitGroup{}
	currentStudentRefIdMap := map[int]struct{}{}
	for _, student := range allRecords {
		wg.Add(1)
		s := student
		srv.executorPool.Submit(func() {
			defer func() {
				if r := recover(); r != nil {
					srv.logger.Error(ctx, "studentService syncStudentsOrUsers syncStudentOrUser panic", nil,
						zap.Any(logger.PanicMessage, r),
						zap.Any("student", *s),
					)
				}
				wg.Done()
			}()

			srv.syncStudentOrUser(ctx, s)
		})

		currentStudentRefIdMap[s.StudentRefId] = struct{}{}
	}

	wg.Wait()

	db := srv.DB.Session()

	isDeleted := false
	poStudentCond := &po.StudentCond{
		IsDeleted: &isDeleted,
	}
	if cond != nil {
		if cond.StudentName != nil {
			poStudentCond.StudentName = *cond.StudentName
		}
		if cond.ParentPhone != nil {
			poStudentCond.ParentPhone = *cond.ParentPhone
		}
	}

	studentRefIdsInDb, err := srv.studentRepo.GetStudentRefIds(ctx, db, poStudentCond)
	if err != nil {
		srv.logger.Error(ctx, "studentService syncStudentsOrUsers GetStudentRefIds", err)
		return
	}

	for _, studentRefIdInDb := range studentRefIdsInDb {
		if _, found := currentStudentRefIdMap[studentRefIdInDb]; !found {
			if err := srv.DeleteStudent(ctx, studentRefIdInDb); err != nil {
				srv.logger.Error(ctx, "studentService syncStudentsOrUsers DeleteStudent", err, zap.Int("record_ref_id", studentRefIdInDb))
			}
		}
	}

	updatedUsers, err := srv.userCommonSrv.DeactivateUserWithNoActiveStudent(ctx)
	if err != nil {
		srv.logger.Error(ctx, "studentService syncStudentsOrUsers userRepo.DeactivateUserWithNoActiveStudent", err)
	}

	if len(updatedUsers) > 0 {
		userId := make([]string, 0, len(updatedUsers))
		for _, updatedUser := range updatedUsers {
			userId = append(userId, fmt.Sprintf("{id: %d, account: %s}", updatedUser.UserId, updatedUser.Account))
		}
		srv.logger.Info(ctx, "studentService syncStudentsOrUsers userRepo.DeactivateUserWithNoActiveStudent updatedUsers", zap.String("userIds", strings.Join(userId, ", ")))
	}
}

func (srv *StudentService) syncStudentOrUser(ctx context.Context, student *bo.Student) {
	boStudent, err := srv.studentCommonSrv.GetStudent(ctx, &bo.StudentCond{StudentRefId: student.StudentRefId})
	if err != nil && !errors.Is(err, errs.DbErr.NoRow) {
		srv.logger.Error(ctx, "studentService syncStudentOrUser studentCommonSrv.GetStudent", err, zap.Int("record_ref_id", student.StudentRefId))
		return
	}

	boUser, err := srv.userCommonSrv.GetUser(ctx, &bo.UserCond{Accounts: []string{student.ParentPhone}})
	if err != nil && !errors.Is(err, errs.DbErr.NoRow) {
		srv.logger.Error(ctx, "studentService syncStudentOrUser userCommonSrv.GetUser", err)
		return
	}

	if boStudent == nil && boUser == nil {
		err := srv.UserRegisterAndCreateStudent(ctx, student)
		if err != nil {
			srv.logger.Error(ctx, "studentService syncStudentOrUser UserRegisterAndCreateStudent", err)
			return
		}
	} else if boUser == nil {
		createUserData, err := srv.genDefaultCreateUserData(ctx, student)
		if err != nil {
			srv.logger.Error(ctx, "studentService syncStudentOrUser genDefaultCreateUserData", err)
			return
		}
		_, err = srv.userCommonSrv.CreateUser(ctx, createUserData)
		if err != nil {
			srv.logger.Error(ctx, "studentService syncStudentOrUser userCommonSrv.CreateUser", err)
			return
		}

		if err := srv.UpdateStudent(ctx, student); err != nil {
			srv.logger.Error(ctx, "studentService syncStudentOrUser boUser == nil UpdateStudent", err)
			return
		}
		//如果在 kintone 上刪除了照理說不會再拿到相同 id 的資料，這裡是為了避免程式邏輯錯誤導致誤刪除了不該刪的記錄
		//所以 API 如果取得到這筆確實存在的記錄但是 DB 卻標示為刪除時，還是重置該記錄的刪除狀態
		if boStudent.IsDeleted || boStudent.DeletedAt != nil {
			if err := srv.studentRepo.ResetFromDeleted(ctx, srv.DB.Session(), (po.Student{}).TableName(), func(db *gorm.DB) *gorm.DB {
				db.Where("student_id = ?", boStudent.StudentId)
				return db
			}); err != nil {
				srv.logger.Error(ctx, "studentService syncStudentOrUser boUser == nil ResetFromDeleted", err)
				return
			}
		}
	} else if boStudent == nil {
		if err := srv.addStudent(ctx, boUser.UserId, student); err != nil {
			srv.logger.Error(ctx, "studentService syncStudentOrUser addStudent", err)
			return
		}
	} else {
		if err := srv.UpdateStudent(ctx, student); err != nil {
			srv.logger.Error(ctx, "studentService syncStudentOrUser UpdateStudent", err)
			return
		}
		//如果在 kintone 上刪除了照理說不會再拿到相同 id 的資料，這裡是為了避免程式邏輯錯誤導致誤刪除了不該刪的記錄
		//所以 API 如果取得到這筆確實存在的記錄但是 DB 卻標示為刪除時，還是重置該記錄的刪除狀態
		if boStudent.IsDeleted || boStudent.DeletedAt != nil {
			if err := srv.studentRepo.ResetFromDeleted(ctx, srv.DB.Session(), (po.Student{}).TableName(), func(db *gorm.DB) *gorm.DB {
				db.Where("student_id = ?", boStudent.StudentId)
				return db
			}); err != nil {
				srv.logger.Error(ctx, "studentService syncStudentOrUser ResetFromDeleted", err)
				return
			}
		}
	}
}
