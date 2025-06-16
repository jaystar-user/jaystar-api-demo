package dto

type AdminGetStudentsIO struct {
	StudentName      *string `form:"student_name"`
	StudentRefId     *int    `form:"student_ref_id"`
	ParentPhone      *string `form:"parent_phone"`
	Mode             *string `form:"mode"`
	IsDeleted        *bool   `form:"is_deleted"`
	IsSettleNormally *bool   `form:"is_settle_normally"`
	*PagerIO
}

type StudentIO struct {
	StudentName *string `form:"student_name"`
	ParentPhone *string `form:"parent_phone"`
}

type StudentVO struct {
	StudentId        string  `json:"student_id"`
	StudentRefId     int     `json:"student_ref_id"`
	StudentName      string  `json:"student_name"`
	ParentName       string  `json:"parent_name"`
	ParentPhone      string  `json:"parent_phone"`
	Balance          float64 `json:"balance"`
	Mode             string  `json:"mode"`
	IsSettleNormally bool    `json:"is_settle_normally"`
}

type AdminStudentVO struct {
	StudentVO
	AdminVO
}

type KintoneWebhookStudentIO struct {
	KintoneWebhookIO
	Record StudentRecord `json:"record"`
}
