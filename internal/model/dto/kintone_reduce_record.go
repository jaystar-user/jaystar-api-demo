package dto

import (
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/bo"
	"jaystar/internal/utils/strUtil"
	"time"

	"golang.org/x/xerrors"
)

type ReduceRecordReq struct {
	StudentName    string    `kQuery:"studentName"`
	Limit          int       `kQuery:"limit"`
	Offset         int       `kQuery:"offset"`
	ClassTimeStart time.Time `kQuery:"classTime >="`
	ClassTimeEnd   time.Time `kQuery:"classTime <="`
}

type ReduceRecordRes struct {
	Records    []ReduceRecordRecord `json:"records"`
	TotalCount string               `json:"totalCount"`
}

type ReduceRecordRecord struct {
	Id           IdField       `json:"$id"`
	StudentName  StringField   `json:"studentName"`
	ClassLevel   StringField   `json:"classLevel"`
	ClassType    StringField   `json:"classType"`
	ClassTime    DateTimeField `json:"classTime"`
	TeacherName  StringField   `json:"teacherName"`
	ReducePoints FloatField    `json:"reducePoints"`
	AttendStatus CheckboxField `json:"attendStatus"`
	Description  StringField   `json:"description"`
	CreatedBy    StringOpField `json:"createdBy"`
	CreatedAt    DateTimeField `json:"createdAt"`
	UpdatedBy    StringOpField `json:"updatedBy"`
	UpdatedAt    DateTimeField `json:"updatedAt"`
}

func (rr *ReduceRecordRecord) ToKintoneReduceRecord() (*bo.KintoneReduceRecord, error) {
	reduceRecord := &bo.KintoneReduceRecord{}
	id, err := rr.Id.ToId()
	if err != nil {
		return nil, xerrors.Errorf("Id.ToId: %w", err)
	}
	reduceRecord.Id = id
	kintoneStudentName := rr.StudentName.ToString()
	reduceRecord.KintoneStudentName = kintoneStudentName
	reduceRecord.StudentName = strUtil.GetStudentNameByStudentName(kintoneStudentName)
	reduceRecord.ParentPhone = strUtil.GetParentPhoneByStudentName(kintoneStudentName)
	reduceRecord.ClassLevel = kintone.ClassLevelToEnum(rr.ClassLevel.ToString())
	reduceRecord.ClassType = kintone.ClassTypeToEnum(rr.ClassType.ToString())
	reduceRecord.ClassTime, err = rr.ClassTime.ToDateTime()
	if err != nil {
		return nil, xerrors.Errorf("ClassTime.ToDateTime: %w", err)
	}
	reduceRecord.TeacherName = rr.TeacherName.ToString()
	reduceRecord.ReducePoints, err = rr.ReducePoints.ToFloat()
	if err != nil {
		return nil, xerrors.Errorf("ReducePoints.ToFloat: %w", err)
	}
	if len(rr.AttendStatus.Value) > 0 && rr.AttendStatus.Value[0] == "出席" {
		reduceRecord.AttendStatus = true
	}
	reduceRecord.CreatedBy = rr.CreatedBy.ToString()
	reduceRecord.CreatedAt, _ = rr.CreatedAt.ToDateTime()
	reduceRecord.UpdatedBy = rr.UpdatedBy.ToString()
	reduceRecord.UpdatedAt, _ = rr.UpdatedAt.ToDateTime()

	return reduceRecord, nil
}

type UpdateReduceRecordsReq struct {
	Records []UpdateReduceRecord `json:"records"`
}

type UpdateReduceRecord struct {
	KintoneUpdateIdBase
	Record UpdateReduceRecordValue `json:"record"`
}

type UpdateReduceRecordValue struct {
	StudentName NormalField `json:"studentName"`
}

type UpdateReduceRecordsRes struct {
	Records []KintoneApiUpdateBaseResponse `json:"records"`
}
