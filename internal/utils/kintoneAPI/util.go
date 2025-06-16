package kintoneAPI

import (
	"errors"
	"fmt"
	kintoneConst "jaystar/internal/constant/kintone"
	"log"
	"reflect"
	"strings"
	"time"
)

const (
	OPERATORS = "=,!=,>,<,>=,in,not in,like,not like"
)

func ParseReqStructToMap(input interface{}) (map[string]interface{}, error) {
	reType := reflect.TypeOf(input)
	if reType.Kind() != reflect.Ptr || reType.Elem().Kind() != reflect.Struct {
		return nil, errors.New("ParseReqToQuery error : neither pointer nor struct")
	}

	reValue := reflect.ValueOf(input).Elem()

	result := make(map[string]interface{})
	for i := 0; i < reValue.NumField(); i++ {
		structField := reValue.Type().Field(i)
		fieldTag := structField.Tag
		kQueryKey := fieldTag.Get("kQuery")

		if kQueryKey == "" {
			log.Printf("ParseReqToQuery cannot get input struct field tag by reflect... %v", kQueryKey)
			continue
		}

		reField := reValue.Field(i)

		switch reField.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			result[kQueryKey] = reField.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			result[kQueryKey] = reField.Uint()
		case reflect.String:
			if reField.IsZero() {
				continue
			}

			result[kQueryKey] = reField.String()
		case reflect.Struct:
			if reField.IsZero() {
				continue
			}

			if reField.Type() == reflect.TypeOf(time.Time{}) {
				result[kQueryKey] = reField.Interface().(time.Time).Format(time.RFC3339)
			}
		}
	}
	return result, nil

}

func ConvMapToQueryStringByAnd(m map[string]interface{}) string {
	var queries string
	var isSet = false
	for k, v := range m {
		if k == kintoneConst.QueryQueryOrderBy || k == kintoneConst.QueryQueryLimit || k == kintoneConst.QueryQueryOffset {
			continue
		}

		isSet = false
		switch val := v.(type) {
		case int64:
			isSet = true
			queries += fmt.Sprintf("%s %d", k, val)
		case uint64:
			isSet = true
			queries += fmt.Sprintf("%s %d", k, val)
		case string:
			isSet = true

			// 检查是否有带上运算符号
			operators := strings.Split(OPERATORS, ",")
			found := false
			for _, op := range operators {
				if strings.Contains(k, op) {
					switch op {
					case "=", "!=", ">", "<", ">=", "like", "not like":
						// 查询字串要加上双引号
						queries += fmt.Sprintf("%s \"%s\"", k, val)
					case "in", "not in":
						// 查询复数字串要个别加上双引号
						valArr := strings.Split(val, ",")
						for k := range valArr {
							valArr[k] = "\"" + valArr[k] + "\""
						}
						queries += fmt.Sprintf("%s (%s)", k, strings.Join(valArr, ","))
					}
					found = true
					break
				}
			}
			if !found {
				// 查询字串要加上双引号
				queries += fmt.Sprintf("%s = \"%s\"", k, val)
			}
		}

		if isSet {
			queries += " and "
		}
	}

	queries = strings.TrimSuffix(queries, " and ")

	if v, found := m[kintoneConst.QueryQueryOrderBy]; found {
		if val, ok := v.(string); ok {
			queries += fmt.Sprintf(" %s %s", kintoneConst.QueryQueryOrderBy, val)
		} else {
			log.Printf("ParseReqToQuery cannot assert correct type for interface value: %v", v)
		}
	}
	if v, found := m[kintoneConst.QueryQueryLimit]; found {
		if val, ok := v.(int64); ok {
			queries += fmt.Sprintf(" %s %d", kintoneConst.QueryQueryLimit, val)
		} else {
			log.Printf("ParseReqToQuery cannot assert correct type for interface value: %v", v)
		}
	}
	if v, found := m[kintoneConst.QueryQueryOffset]; found {
		if val, ok := v.(int64); ok {
			queries += fmt.Sprintf(" %s %d", kintoneConst.QueryQueryOffset, val)
		} else {
			log.Printf("ParseReqToQuery cannot assert correct type for interface value: %v", v)
		}
	}

	return queries
}
