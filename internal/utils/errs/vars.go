package errs

import (
	"github.com/SeanZhenggg/go-utils/errortool"
)

var (
	Define = errortool.ProvideDefine().
		SetErrCodeOptions(errortool.ErrCodeOptions{
			Min:   1,
			Max:   9999,
			Range: 10000,
		})
	CommonErr       = ProvideCommonError()
	DbErr           = ProvideDBError()
	UserErr         = ProvideUserSrvError()
	StudentErr      = ProvideStudentSrvError()
	ReduceRecordErr = ProvideReduceRecordError()
	ScheduleErr     = ProvideScheduleError()
	KintoneErr      = ProvideKintoneError()
)
