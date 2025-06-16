package kintone

type Dto struct {
	Key   string
	Value string
}

type ChargingStatus int

const (
	NotCharged ChargingStatus = iota
	IsCharged
)

func ChargingStatusToEnum(raw string) ChargingStatus {
	switch raw {
	case "已入款":
		return IsCharged
	default:
		return NotCharged
	}
}

func (cs ChargingStatus) ToValue() bool {
	switch cs {
	case IsCharged:
		return true
	default:
		return false
	}
}

func (cs ChargingStatus) ToName() string {
	switch cs {
	case IsCharged:
		return "已入款"
	default:
		return "未入款"
	}
}

type ChargingMethod int

func (cm ChargingMethod) ToKey() string {
	return ChargingMethodMap[cm].Key
}

func (cm ChargingMethod) ToValue() string {
	return ChargingMethodMap[cm].Value
}

const (
	ChargingMethodMismatched ChargingMethod = iota
	Cash
	Atm
	CreditCard
)

var ChargingMethodMap = map[ChargingMethod]Dto{
	ChargingMethodMismatched: {Key: "mismatched", Value: "未匹配"},
	Cash:                     {Key: "cash", Value: "現金"},
	Atm:                      {Key: "atm", Value: "匯款ATM"},
	CreditCard:               {Key: "credit_card", Value: "刷卡"},
}

func ChargingMethodToEnum(raw []string) []ChargingMethod {
	res := make([]ChargingMethod, 0, len(raw))
	for _, v := range raw {
		switch v {
		case "現金", "cash":
			res = append(res, Cash)
		case "匯款ATM", "atm":
			res = append(res, Atm)
		case "刷卡", "credit_card":
			res = append(res, CreditCard)
		default:
			res = append(res, ChargingMethodMismatched)
		}
	}
	return res
}

func ChargingMethodToKey(enums []ChargingMethod) []string {
	res := make([]string, 0, len(enums))
	for _, v := range enums {
		res = append(res, v.ToKey())
	}
	return res
}

func ChargingMethodToValue(enums []ChargingMethod) []string {
	res := make([]string, 0, len(enums))
	for _, v := range enums {
		res = append(res, v.ToValue())
	}
	return res
}

type ClassType int

const (
	ClassTypeMismatched ClassType = iota
	OneToOne
	OneToTwo
	OneToThree
	Group
)

var ClassTypeMap = map[ClassType]Dto{
	ClassTypeMismatched: {Key: "mismatched", Value: "未匹配"},
	OneToOne:            {Key: "one_to_one", Value: "一對一"},
	OneToTwo:            {Key: "one_to_two", Value: "一對二"},
	OneToThree:          {Key: "one_to_three", Value: "一對三"},
	Group:               {Key: "group", Value: "團課"},
}

func (ct ClassType) ToKey() string {
	return ClassTypeMap[ct].Key
}

func (ct ClassType) ToValue() string {
	return ClassTypeMap[ct].Value
}

func ClassTypeToEnum(raw string) ClassType {
	switch raw {
	case "one_to_one", "一對一":
		return OneToOne
	case "one_to_two", "一對二":
		return OneToTwo
	case "one_to_three", "一對三":
		return OneToThree
	case "group", "團課":
		return Group
	default:
		return ClassTypeMismatched
	}
}

type ClassLevel int

const (
	ClassLevelMismatched ClassLevel = iota
	L0
	L1
	L2
	L3
	L4
)

var ClassLevelMap = map[ClassLevel]Dto{
	ClassLevelMismatched: {Key: "mismatched", Value: "未匹配"},
	L0:                   {Key: "L0", Value: "寶寶班"},
	L1:                   {Key: "L1", Value: "Level 1"},
	L2:                   {Key: "L2", Value: "Level 2"},
	L3:                   {Key: "L3", Value: "Level 3"},
	L4:                   {Key: "L4", Value: "Level 4"},
}

func (cl ClassLevel) ToKey() string {
	return ClassLevelMap[cl].Key
}

func (cl ClassLevel) ToValue() string {
	return ClassLevelMap[cl].Value
}

func ClassLevelToEnum(raw string) ClassLevel {
	switch raw {
	case "L0", "寶寶班":
		return L0
	case "L1", "Level 1":
		return L1
	case "L2", "Level 2":
		return L2
	case "L3", "Level 3":
		return L3
	case "L4", "Level 4":
		return L4
	default:
		return ClassLevelMismatched
	}
}

type WebhookType string

const (
	AddRecordType    WebhookType = "ADD_RECORD"
	UpdateRecordType WebhookType = "UPDATE_RECORD"
	UpdateStatusType WebhookType = "UPDATE_STATUS"
	DeleteRecordType WebhookType = "DELETE_RECORD"
)

type Mode int

const (
	ModeMismatched Mode = iota
	ModeRecharge
	ModeSemester
)

var ModeMap = map[Mode]Dto{
	ModeMismatched: {Key: "mismatched", Value: "未匹配"},
	ModeRecharge:   {Key: "recharge", Value: "舊制"},
	ModeSemester:   {Key: "semester", Value: "學期制"},
}

func ModeToEnum(raw string) Mode {
	switch raw {
	case "recharge", "舊制":
		return ModeRecharge
	case "semester", "學期制":
		return ModeSemester
	default:
		return ModeMismatched
	}
}

func (m Mode) ToValue() string {
	return ModeMap[m].Value
}

func (m Mode) ToKey() string {
	return ModeMap[m].Key
}

func StringToBool(raw string) bool {
	switch raw {
	case "是":
		return true
	default:
		return false
	}
}
