package dto

type SyncAllByStudentIO struct {
	StudentName string `json:"student_name" binding:"required"`
	ParentPhone string `json:"parent_phone" binding:"required"`
}
