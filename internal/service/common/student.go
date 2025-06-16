package common

import (
	"context"
	"golang.org/x/xerrors"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/database"
	"jaystar/internal/interfaces"
	"jaystar/internal/model/bo"
	"jaystar/internal/model/dto"
	"jaystar/internal/model/po"
	"jaystar/internal/utils/errs"
	"reflect"
	"strconv"
)

func ProvideStudentCommonService(
	db database.IPostgresDB,
	studentRepo interfaces.IStudentRepo,
	userRepo interfaces.IUserRepo,
	kintoneStudentRepo interfaces.IKintoneStudentRepo,
) *StudentCommonService {
	return &StudentCommonService{
		DB:                 db,
		studentRepo:        studentRepo,
		userRepo:           userRepo,
		kintoneStudentRepo: kintoneStudentRepo,
	}
}

type StudentCommonService struct {
	DB                 database.IPostgresDB
	studentRepo        interfaces.IStudentRepo
	userRepo           interfaces.IUserRepo
	kintoneStudentRepo interfaces.IKintoneStudentRepo
}

func (srv *StudentCommonService) GetKintoneStudents(ctx context.Context, cond *dto.StudentReq) ([]*bo.Student, int, error) {
	studentRes, err := srv.kintoneStudentRepo.GetKintoneStudents(ctx, cond)
	if err != nil {
		return nil, 0, xerrors.Errorf("studentCommonService GetKintoneStudent kintoneStudentRepo.GetKintoneStudents: %w", err)
	}

	total, err := strconv.ParseInt(studentRes.TotalCount, 10, 64)
	if err != nil {
		return nil, 0, xerrors.Errorf("studentCommonService GetKintoneStudents strconv.ParseInt: %w", err)
	}

	boStudents := make([]*bo.Student, 0, len(studentRes.Records))
	for _, record := range studentRes.Records {
		boStudent, err := record.ToStudent()
		if err != nil {
			return nil, 0, xerrors.Errorf("studentCommonService GetKintoneStudent poStudentRes.ToStudent: %w", err)
		}

		boStudents = append(boStudents, boStudent)
	}

	return boStudents, int(total), nil
}

func (srv *StudentCommonService) GetStudent(ctx context.Context, cond *bo.StudentCond) (*bo.Student, error) {
	if reflect.ValueOf(cond).Elem().IsZero() {
		return nil, errs.CommonErr.RequestParamError
	}

	isDeleted := false
	poStudentCond := &po.StudentCond{
		IsDeleted: &isDeleted,
	}
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

	db := srv.DB.Session()
	poStudent, err := srv.studentRepo.GetStudent(ctx, db, poStudentCond)
	if err != nil {
		return nil, xerrors.Errorf("studentRepo.GetStudent: %w", err)
	}

	boStudent := &bo.Student{
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
	}

	return boStudent, nil
}
