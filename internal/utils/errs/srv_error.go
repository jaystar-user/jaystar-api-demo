package errs

const (
	UserGroupCode int = iota + 2
	StudentGroupCode
	ReduceRecordGroupCode
	ScheduleGroupCode
	KintoneGroupCode
)

func ProvideUserSrvError() *userSrvError {
	group := Define.GenErrorGroup(UserGroupCode)

	return &userSrvError{
		AccOrPwdVerificationFailedErr: group.GenError(1, "帳號或密碼錯誤"),
		AccountDeactivatedErr:         group.GenError(2, "帳號已被停用"),
		AccountOrPasswordInvalidErr:   group.GenError(3, "無效的帳號或密碼"),
		UpdatePasswordEmptyError:      group.GenError(4, "密碼欄位不得為空"),
		UserNotFoundErr:               group.GenError(5, "找不到對應的使用者"),
		UserIdInvalidErr:              group.GenError(6, "無效的使用者ID"),
		AccountDuplicatedErr:          group.GenError(7, "使用者帳號重複"),
	}
}

type userSrvError struct {
	AccOrPwdVerificationFailedErr error
	AccountDeactivatedErr         error
	AccountOrPasswordInvalidErr   error
	UserIdInvalidErr              error
	UpdatePasswordEmptyError      error
	UserNotFoundErr               error
	AccountDuplicatedErr          error
}

func ProvideStudentSrvError() *studentSrvError {
	group := Define.GenErrorGroup(StudentGroupCode)

	return &studentSrvError{
		GetKintoneStudentNotFoundErr:         group.GenError(1, "找不到學生資料"),
		StudentIdInvalidErr:                  group.GenError(2, "無效的學生ID資料"),
		StudentNameInvalidErr:                group.GenError(3, "無效的學生姓名資料"),
		ParentPhoneInvalidErr:                group.GenError(4, "無效的家長電話資料"),
		StudentNameAndParentPhoneRequiredErr: group.GenError(5, "學生姓名與家長電話必填"),
		StudentNotFoundErr:                   group.GenError(6, "沒有找到對應的學生資料"),
	}
}

type studentSrvError struct {
	GetKintoneStudentNotFoundErr         error
	StudentIdInvalidErr                  error
	StudentNameInvalidErr                error
	ParentPhoneInvalidErr                error
	StudentNameAndParentPhoneRequiredErr error
	StudentNotFoundErr                   error
}

func ProvideReduceRecordError() *reduceRecordError {
	group := Define.GenErrorGroup(ReduceRecordGroupCode)

	return &reduceRecordError{
		InvalidStudentNameErr: group.GenError(1, "學生姓名欄位資料解析錯誤"),
	}
}

type reduceRecordError struct {
	InvalidStudentNameErr error
}

func ProvideScheduleError() *scheduleError {
	group := Define.GenErrorGroup(ScheduleGroupCode)

	return &scheduleError{
		InvalidAttendanceDataError: group.GenError(1, "無效的出席資料"),
		InvalidStudentNameErr:      group.GenError(2, "學生姓名欄位資料解析錯誤"),
		InvalidScheduleRefIdError:  group.GenError(3, "無效的課表ID資料"),
	}
}

type scheduleError struct {
	InvalidAttendanceDataError error
	InvalidStudentNameErr      error
	InvalidScheduleRefIdError  error
}

func ProvideKintoneError() *kintoneError {
	group := Define.GenErrorGroup(KintoneGroupCode)

	return &kintoneError{
		ResponseEmptyError: group.GenError(1, "查詢不到 kintone 資料"),
	}
}

type kintoneError struct {
	ResponseEmptyError error
}
