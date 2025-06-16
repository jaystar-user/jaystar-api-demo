package dto

import (
	"golang.org/x/xerrors"
	"jaystar/internal/model/bo"
	"jaystar/internal/utils/strUtil"
)

type GetPointCardReq struct {
	StudentName string `kQuery:"studentName"`
	Limit       int    `kQuery:"limit"`
	Offset      int    `kQuery:"offset"`
}

type GetPointCardRes struct {
	Records    []PointCardRecord `json:"records"`
	TotalCount string            `json:"totalCount"`
}

type PointCardRecord struct {
	Id          IdField       `json:"$id"`
	StudentName StringField   `json:"studentName"`
	RestPoints  FloatField    `json:"restPoints"`
	CreatedBy   StringOpField `json:"建立人"`
	CreatedAt   DateTimeField `json:"建立時間"`
	UpdatedBy   StringOpField `json:"更新人"`
	UpdatedAt   DateTimeField `json:"更新時間"`
}

func (rr *PointCardRecord) ToPointCard() (*bo.PointCard, error) {
	pointCard := &bo.PointCard{}
	id, err := rr.Id.ToId()
	if err != nil {
		return nil, xerrors.Errorf("Id.ToId: %w", err)
	}
	pointCard.RecordRefId = id
	kintoneStudentName := rr.StudentName.ToString()
	pointCard.KintoneStudentName = kintoneStudentName
	pointCard.StudentName = strUtil.GetStudentNameByStudentName(kintoneStudentName)
	pointCard.ParentPhone = strUtil.GetParentPhoneByStudentName(kintoneStudentName)
	pointCard.RestPoints, err = rr.RestPoints.ToFloat()
	if err != nil {
		return nil, xerrors.Errorf("RestPoints.ToFloat: %w", err)
	}

	return pointCard, nil
}

type UpdatePointCardReq struct {
	KintoneUpdateIdBase
	Record UpdatePointCardRecord `json:"record"`
}

type UpdatePointCardRecord struct {
	StudentName NormalField `json:"studentName"`
}

type UpdatePointCardRes struct {
	KintoneApiUpdateBaseResponse
}

type UpdatePointCardsReq struct {
	Records []UpdatePointCardsRecord `json:"records"`
}

type UpdatePointCardsRecord struct {
	UpdateKey struct {
		Field string `json:"field"`
		Value string `json:"value"`
	} `json:"updateKey"`
	Record UpdatePointCardsRecordValue `json:"record"`
}

type UpdatePointCardsRecordValue struct {
	ClearPoints NormalField `json:"clearPoints"`
}

type UpdatePointCardsRes struct {
	Records []KintoneApiUpdateBaseResponse `json:"records"`
}
