package dto

import (
	"golang.org/x/xerrors"
	"jaystar/internal/constant/kintone"
	"jaystar/internal/model/bo"
)

type StudentReq struct {
	StudentName string `kQuery:"studentName"`
	ParentPhone string `kQuery:"parentPhone"`
	Limit       int    `kQuery:"limit"`
	Offset      int    `kQuery:"offset"`
}

type StudentRes struct {
	Records    []StudentRecord `json:"records"`
	TotalCount string          `json:"totalCount"`
}

type StudentRecord struct {
	Id               IdField       `json:"$id"`
	StudentName      StringField   `json:"studentName"`
	Birthday         StringField   `json:"birthday"`
	Gender           StringField   `json:"gender"`
	ParentName       StringField   `json:"parentName"`
	ParentPhone      StringField   `json:"parentPhone"`
	RestPoints       FloatField    `json:"restPoints"`
	Description      StringField   `json:"description"`
	TaxId            StringField   `json:"taxId"`
	Relationship     StringField   `json:"relationship"`
	Mode             StringField   `json:"mode"`
	IsSettleNormally StringField   `json:"isSettleNormally"`
	CreatedBy        StringOpField `json:"createdBy"`
	CreatedAt        DateTimeField `json:"createdAt"`
	UpdatedBy        StringOpField `json:"updatedBy"`
	UpdatedAt        DateTimeField `json:"updatedAt"`
}

func (s *StudentRecord) ToStudent() (*bo.Student, error) {
	boKintoneStudent := &bo.Student{}
	id, err := s.Id.ToId()
	if err != nil {
		return nil, xerrors.Errorf("bo student StudentRecord ToStudent s.Id.ToId: %w", err)
	}
	boKintoneStudent.StudentRefId = id
	boKintoneStudent.StudentName = s.StudentName.ToString()
	boKintoneStudent.ParentName = s.ParentName.ToString()
	boKintoneStudent.ParentPhone = s.ParentPhone.ToString()
	boKintoneStudent.Mode = kintone.ModeToEnum(s.Mode.ToString())
	boKintoneStudent.IsSettleNormally = kintone.StringToBool(s.IsSettleNormally.ToString())
	boKintoneStudent.CreatedAt, _ = s.CreatedAt.ToDateTime()
	boKintoneStudent.UpdatedAt, _ = s.UpdatedAt.ToDateTime()

	return boKintoneStudent, nil
}
