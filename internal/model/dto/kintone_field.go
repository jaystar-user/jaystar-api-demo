package dto

import (
	"strconv"
	"time"
)

type IdField struct {
	NormalField
}

func (f IdField) ToId() (int, error) {
	val, err := strconv.Atoi(f.Value)
	if err != nil {
		return 0, err
	}

	return val, nil
}

type DateTimeField struct {
	NormalField
}

func (f DateTimeField) ToDate() (time.Time, error) {
	return time.Parse(time.DateOnly, f.Value)
}

func (f DateTimeField) ToDateTime() (time.Time, error) {
	return time.Parse(time.RFC3339, f.Value)
}

type IntField struct {
	NormalField
}

func (f IntField) ToInt() (int, error) {
	return strconv.Atoi(f.Value)
}

type FloatField struct {
	NormalField
}

func (f FloatField) ToFloat() (float64, error) {
	return strconv.ParseFloat(f.Value, 64)

}

type StringField struct {
	NormalField
}

func (f StringField) ToString() string {
	return f.Value
}

type StringOpField struct {
	OperatorField
}

func (f StringOpField) ToString() string {
	return f.Value.Name
}

type NormalField struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type OperatorField struct {
	Type  string        `json:"type"`
	Value CodeNameValue `json:"value"`
}

type CheckboxField struct {
	Type  string   `json:"type"`
	Value []string `json:"value"`
}

type CodeNameValue struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
