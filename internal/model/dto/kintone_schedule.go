package dto

import (
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/bo"
	"jaystar/internal/utils/errs"
	"jaystar/internal/utils/strUtil"
	"time"

	"golang.org/x/xerrors"
)

type ScheduleReq struct {
	StudentName    string    `kQuery:"studentName in"`
	ClassTimeStart time.Time `kQuery:"classTime >="`
	ClassTimeEnd   time.Time `kQuery:"classTime <="`
	Limit          int       `kQuery:"limit"`
	Offset         int       `kQuery:"offset"`
	OrderBy        string    `kQuery:"order by"`
}

type ScheduleRes struct {
	Records    []ScheduleRecord `json:"records"`
	TotalCount string           `json:"totalCount"`
}

type ScheduleRecord struct {
	Id          IdField       `json:"$id"`
	TeacherName StringField   `json:"teacherName"`
	ClassLevel  StringField   `json:"classLevel"`
	ClassType   StringField   `json:"classType"`
	ClassTime   DateTimeField `json:"classTime"`
	Attendance  Attendance    `json:"attendance"`
	Description StringField   `json:"description"`
	CreatedBy   StringOpField `json:"createdBy"`
	CreatedAt   DateTimeField `json:"createdAt"`
	UpdatedBy   StringOpField `json:"updatedBy"`
	UpdatedAt   DateTimeField `json:"updatedAt"`
}

func (s *ScheduleRecord) ToSchedules() ([]*bo.KintoneSchedule, error) {
	if len(s.Attendance.Value) == 0 {
		return nil, errs.ScheduleErr.InvalidAttendanceDataError
	}
	boSchedules := make([]*bo.KintoneSchedule, 0, len(s.Attendance.Value))
	id, err := s.Id.ToId()
	if err != nil {
		return nil, xerrors.Errorf("Id.ToId: %w", err)
	}

	for _, record := range s.Attendance.Value {
		kintoneStudentName := record.Value.StudentName.ToString()
		boSchedule := &bo.KintoneSchedule{
			ScheduleRefId:      id,
			KintoneStudentName: kintoneStudentName,
			StudentName:        strUtil.GetStudentNameByStudentName(kintoneStudentName),
			ParentPhone:        strUtil.GetParentPhoneByStudentName(kintoneStudentName),
			TeacherName:        s.TeacherName.ToString(),
			ClassType:          kintone.ClassTypeToEnum(s.ClassType.Value),
			ClassLevel:         kintone.ClassLevelToEnum(s.ClassLevel.Value),
			Description:        s.Description.ToString(),
			CreatedBy:          s.CreatedBy.ToString(),
			UpdatedBy:          s.UpdatedBy.ToString(),
		}

		boSchedule.ClassTime, err = s.ClassTime.ToDateTime()
		if err != nil {
			return nil, xerrors.Errorf("ClassTime.ToDateTime: %w", err)
		}
		boSchedule.CreatedAt, _ = s.CreatedAt.ToDateTime()
		boSchedule.UpdatedAt, _ = s.UpdatedAt.ToDateTime()

		if record.Value.RecordId.Value != "" {
			recordId, err := record.Value.RecordId.ToId()
			if err != nil {
				return nil, xerrors.Errorf("record.Value.RecordId.ToId: %w", err)
			}
			boSchedule.RecordRefId = &recordId
		}

		boSchedules = append(boSchedules, boSchedule)
	}

	return boSchedules, nil
}

type UpdateSchedulesReq struct {
	Records []UpdateSchedule `json:"records"`
}

type UpdateSchedule struct {
	KintoneUpdateIdBase
	Record UpdateScheduleValue `json:"record"`
}

type UpdateScheduleValue struct {
	Attendance Attendance `json:"attendance"`
}

type Attendance struct {
	Type  string             `json:"type"`
	Value []*AttendanceValue `json:"value"`
}

type AttendanceValue struct {
	Id    string                `json:"id"`
	Value AttendanceRecordValue `json:"value"`
}

type AttendanceRecordValue struct {
	RecordId     IdField       `json:"recordId"`
	AttendStatus CheckboxField `json:"attendStatus"`
	StudentName  StringField   `json:"studentName"`
	ReducePoints FloatField    `json:"reducePoints"`
}

type UpdateSchedulesRes struct {
	Records []KintoneApiUpdateBaseResponse `json:"records"`
}
