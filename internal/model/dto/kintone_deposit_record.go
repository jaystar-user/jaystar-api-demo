package dto

import (
	"golang.org/x/xerrors"
	kintoneConst "jaystar/internal/constant/kintone"
	"jaystar/internal/model/bo"
	"jaystar/internal/utils/strUtil"
	"time"
)

type DepositRecordReq struct {
	StudentName       string    `kQuery:"studentName"`
	ChargingDateStart time.Time `kQuery:"chargingDate >="`
	ChargingDateEnd   time.Time `kQuery:"chargingDate <="`
	Limit             int       `kQuery:"limit"`
	Offset            int       `kQuery:"offset"`
}

type DepositRecordRes struct {
	Records    []KintoneDepositRecordRecordDto `json:"records"`
	TotalCount string                          `json:"totalCount"`
}

type KintoneDepositRecordRecordDto struct {
	Id                   IdField       `json:"$id"`
	ChargingMethod       CheckboxField `json:"chargingMethod"`
	ActualChargingAmount IntField      `json:"actualChargingAmount"`
	Gender               StringField   `json:"gender"`
	DepositedPoints      IntField      `json:"depositedPoints"`
	TeacherName          StringField   `json:"teacherName"`
	Description          StringField   `json:"description"`
	ParentPhone          StringField   `json:"parentPhone"` // deprecated
	ChargingDate         DateTimeField `json:"chargingDate"`
	ChargingAmount       IntField      `json:"chargingAmount"`
	ParentName           StringField   `json:"parentName"` // deprecated
	AccountLastFiveYards StringField   `json:"accountLastFiveYards"`
	TaxId                StringField   `json:"taxId"`
	StudentName          StringField   `json:"studentName"`
	ChargingStatus       StringField   `json:"chargingStatus"`
	CreatedBy            StringOpField `json:"createdBy"`
	CreatedAt            DateTimeField `json:"createdAt"`
	UpdatedBy            StringOpField `json:"updatedBy"`
	UpdatedAt            DateTimeField `json:"updatedAt"`
}

func (d *KintoneDepositRecordRecordDto) ToKintoneDepositRecordBo() (*bo.KintoneDepositRecord, error) {
	boDepositRecord := &bo.KintoneDepositRecord{}
	id, err := d.Id.ToId()
	if err != nil {
		return nil, xerrors.Errorf("d.Id.ToId: %w", err)
	}
	boDepositRecord.Id = id
	kintoneStudentName := d.StudentName.ToString()
	boDepositRecord.KintoneStudentName = kintoneStudentName
	boDepositRecord.StudentName = strUtil.GetStudentNameByStudentName(kintoneStudentName)
	boDepositRecord.ParentPhone = strUtil.GetParentPhoneByStudentName(kintoneStudentName)
	chargingMethod := kintoneConst.ChargingMethodToEnum(d.ChargingMethod.Value)
	boDepositRecord.ChargingMethod = chargingMethod
	boDepositRecord.ChargingStatus = kintoneConst.ChargingStatusToEnum(d.ChargingStatus.ToString())
	actualChargingAmount, err := d.ActualChargingAmount.ToInt()
	if err != nil {
		return nil, xerrors.Errorf("ActualChargingAmount.ToInt: %w", err)
	}
	boDepositRecord.ActualChargingAmount = actualChargingAmount
	boDepositRecord.Gender = d.Gender.ToString()
	depositedPoints, err := d.DepositedPoints.ToInt()
	if err != nil {
		return nil, xerrors.Errorf("DepositedPoints.ToInt: %w", err)
	}
	boDepositRecord.DepositedPoints = depositedPoints
	boDepositRecord.TeacherName = d.TeacherName.ToString()
	boDepositRecord.Description = d.Description.ToString()
	chargingDate, err := d.ChargingDate.ToDate()
	if err != nil {
		return nil, xerrors.Errorf("ChargingDate.ToDate: %w", err)
	}
	boDepositRecord.ChargingDate = chargingDate
	chargingAmount, err := d.ChargingAmount.ToInt()
	if err != nil {
		return nil, xerrors.Errorf("ChargingAmount.ToInt: %w", err)
	}
	boDepositRecord.ChargingAmount = chargingAmount
	boDepositRecord.ParentName = d.ParentName.ToString()
	boDepositRecord.AccountLastFiveYards = d.AccountLastFiveYards.ToString()
	boDepositRecord.TaxId = d.TaxId.ToString()
	boDepositRecord.CreatedBy = d.CreatedBy.ToString()
	createdAt, _ := d.CreatedAt.ToDateTime()
	boDepositRecord.CreatedAt = createdAt
	boDepositRecord.UpdatedBy = d.UpdatedBy.ToString()
	updatedAt, _ := d.UpdatedAt.ToDateTime()
	boDepositRecord.UpdatedAt = updatedAt

	return boDepositRecord, nil
}

type UpdateDepositRecordsReq struct {
	Records []UpdateDepositRecord `json:"records"`
}

type UpdateDepositRecord struct {
	KintoneUpdateIdBase
	Record UpdateDepositRecordValue `json:"record"`
}

type UpdateDepositRecordValue struct {
	StudentName NormalField `json:"studentName"`
}

type UpdateDepositRecordsRes struct {
	Records []KintoneApiUpdateBaseResponse `json:"records"`
}
