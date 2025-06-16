package dto

import (
	"golang.org/x/xerrors"
	"jaystar/internal/model/bo"
	"jaystar/internal/utils/strUtil"
	"time"
)

type SemesterSettleRecordReq struct {
	StudentName string    `kQuery:"studentName"`
	StartTime   time.Time `kQuery:"startTime >="`
	EndTime     time.Time `kQuery:"endTime <="`
	Limit       int       `kQuery:"limit"`
	Offset      int       `kQuery:"offset"`
	OrderBy     string    `kQuery:"order by"`
}

type SemesterSettleRecordRes struct {
	Records    []SemesterSettleRecord `json:"records"`
	TotalCount string                 `json:"totalCount"`
}

type SemesterSettleRecord struct {
	Id          IdField       `json:"$id"`
	StudentName StringField   `json:"studentName"`
	StartTime   DateTimeField `json:"startTime"`
	EndTime     DateTimeField `json:"endTime"`
	ClearPoints FloatField    `json:"clearPoints"`
	CreatedBy   StringOpField `json:"建立人"`
	CreatedAt   DateTimeField `json:"建立時間"`
	UpdatedBy   StringOpField `json:"更新人"`
	UpdatedAt   DateTimeField `json:"更新時間"`
}

func (rr *SemesterSettleRecord) ToSemesterSettleRecord() (*bo.SemesterSettleRecord, error) {
	semesterSettleRecord := &bo.SemesterSettleRecord{}
	id, err := rr.Id.ToId()
	if err != nil {
		return nil, xerrors.Errorf("Id.ToId: %w", err)
	}
	semesterSettleRecord.RecordRefId = id
	kintoneStudentName := rr.StudentName.ToString()
	semesterSettleRecord.KintoneStudentName = kintoneStudentName
	semesterSettleRecord.StudentName = strUtil.GetStudentNameByStudentName(kintoneStudentName)
	semesterSettleRecord.ParentPhone = strUtil.GetParentPhoneByStudentName(kintoneStudentName)
	semesterSettleRecord.StartTime, err = rr.StartTime.ToDate()
	if err != nil {
		return nil, xerrors.Errorf("StartTime.ToDate: %w", err)
	}
	semesterSettleRecord.EndTime, err = rr.EndTime.ToDate()
	if err != nil {
		return nil, xerrors.Errorf("EndTime.ToDate: %w", err)
	}
	semesterSettleRecord.ClearPoints, err = rr.ClearPoints.ToFloat()
	if err != nil {
		return nil, xerrors.Errorf("ClearPoints.ToFloat: %w", err)
	}
	semesterSettleRecord.CreatedBy = rr.CreatedBy.ToString()
	semesterSettleRecord.CreatedAt, _ = rr.CreatedAt.ToDateTime()
	semesterSettleRecord.UpdatedBy = rr.UpdatedBy.ToString()
	semesterSettleRecord.UpdatedAt, _ = rr.UpdatedAt.ToDateTime()

	return semesterSettleRecord, nil
}

type InsertSemesterSettleRecordsReq struct {
	Records []InsertSemesterSettleRecord `json:"records"`
}

type InsertSemesterSettleRecord struct {
	StudentName NormalField `json:"studentName"`
	StartTime   NormalField `json:"startTime"`
	EndTime     NormalField `json:"endTime"`
	ClearPoints NormalField `json:"clearPoints"`
}

type InsertSemesterSettleRecordsRes struct {
	Ids       []string `json:"ids"`
	Revisions []string `json:"revisions"`
}
