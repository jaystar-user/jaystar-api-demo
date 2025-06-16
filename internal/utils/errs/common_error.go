package errs

const (
	CommonGroupCode int = 1
)

func ProvideCommonError() *commonError {
	group := Define.GenErrorGroup(CommonGroupCode)

	return &commonError{
		UnknownError:      group.GenError(1, "伺服器錯誤"),
		RequestParamError: group.GenError(2, "請求參數錯誤"),
		AuthFailedError:   group.GenError(3, "登入驗證失敗"),
		AuthDeniedError:   group.GenError(4, "操作權限不足"),
		CondError:         group.GenError(5, "無效參數"),
	}
}

type commonError struct {
	UnknownError      error
	RequestParamError error
	AuthFailedError   error
	AuthDeniedError   error
	CondError         error
}
