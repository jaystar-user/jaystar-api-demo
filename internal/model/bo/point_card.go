package bo

type PointCard struct {
	RecordId           int64
	RecordRefId        int
	StudentId          int64
	KintoneStudentName string
	StudentName        string
	ParentPhone        string
	RestPoints         float64
}

type GetPointCardCond struct {
	StudentIds []int64
}

type UpdatePointCardCond struct {
	RecordRefId int
}

type UpdatePointCardRecordData struct {
	RestPoints *float64
}

type SyncPointCardCond struct {
	StudentName *string
	ParentPhone *string
	Limit       *int
	Offset      *int
}

type SyncSettledStudentPointCardData struct {
	StudentId          int64
	KintoneStudentName string
	ClearPoints        float64
	RestPoints         float64
}
