package util

import (
	"jaystar/internal/constant/request"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"golang.org/x/xerrors"

	"github.com/SeanZhenggg/go-utils/logger"
	"jaystar/internal/model/dto"
)

type IRequestParse interface {
	Bind(ctx *gin.Context, obj interface{}) error
}

func ProviderRequestParse(logger logger.ILogger) IRequestParse {
	return &requestParseHandler{Logger: logger}
}

type requestParseHandler struct {
	Logger logger.ILogger
}

func (r requestParseHandler) Bind(ctx *gin.Context, obj interface{}) error {
	switch ctx.Request.Method {
	case http.MethodGet:
		return r.bindMethodGet(ctx, obj)
	default:
		if err := ctx.ShouldBindJSON(&obj); err != nil {
			return xerrors.Errorf("requestParseHandler ShouldBindJSON: %w", err)
		}
		return nil
	}
}

func (r requestParseHandler) bindMethodGet(ctx *gin.Context, obj interface{}) error {
	if err := ctx.ShouldBindQuery(obj); err != nil {
		return xerrors.Errorf("requestParseHandler bindMethodGet ShouldBindQuery: %w", err)
	}

	// 用來判斷 request 是否有帶入 *PagerIO 資料
	// NOTE: dto 那邊也需要用指標
	var pager *dto.PagerIO
	// 取得空的 pager 指標資料
	valPagerType := reflect.ValueOf(&pager).Elem().Type()
	// 取得 request 的 reflect.Value 資料
	val := reflect.Indirect(reflect.ValueOf(obj).Elem())
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		// 當發現有Pager Type 時，解析 Pager
		if field.Type() == valPagerType {
			// Pager 是空的，給預設值
			if field.IsNil() {
				p := &dto.PagerIO{Index: request.DefaultIdx, Size: request.DefaultSize}
				field.Set(reflect.ValueOf(&p).Elem())
				break
			}

			pageVal := reflect.Indirect(field)
			pageTyp := pageVal.Type()

			// 判斷 Pager 資料並用預設值覆蓋 zero value
			for idx := 0; idx < pageVal.NumField(); idx++ {
				fieldValue := pageVal.Field(idx)
				fieldName := pageTyp.Field(idx).Name

				switch fieldName {
				case "Index":
					// 利用 type assertion 確定欄位值為 int 型別
					if index, ok := fieldValue.Interface().(int); ok && index == 0 {
						fieldValue.Set(reflect.ValueOf(request.DefaultIdx))
					}
				case "Size":
					// 利用 type assertion 確定欄位值為 int 型別
					if size, ok := fieldValue.Interface().(int); ok && size == 0 {
						fieldValue.Set(reflect.ValueOf(request.DefaultSize))
					}
				}
			}

			break
		}
	}

	return nil
}
