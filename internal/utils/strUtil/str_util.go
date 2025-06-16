package strUtil

import "strings"

// GetAccArrFromAccStr 根據 "," 切分使用者帳號
func GetAccArrFromAccStr(accstr string) []string {
	accsStr := strings.Split(accstr, ",")
	accArr := make([]string, 0, len(accsStr))
	for _, acc := range accsStr {
		accArr = append(accArr, acc)
	}
	return accArr
}

// GetParentPhoneByStudentName 根據學生姓名資料取得家長電話
func GetParentPhoneByStudentName(studentName string) string {
	if len(studentName) == 0 {
		return ""
	}

	tuples := strings.Split(studentName, "/")
	if len(tuples) != 2 {
		return ""
	}

	phone := tuples[1]
	if phone == "" {
		return ""
	}

	return phone
}

// GetStudentNameByStudentName 根據學生姓名資料取得學生姓名
func GetStudentNameByStudentName(studentName string) string {
	if len(studentName) == 0 {
		return ""
	}

	tuples := strings.Split(studentName, "/")
	if len(tuples) != 2 {
		return studentName
	}

	return tuples[0]
}

// IsValidStudentName 確認學生姓名資料是否合法
func IsValidStudentName(studentName string) bool {
	if len(studentName) == 0 {
		return false
	}
	tuples := strings.Split(studentName, "/")

	return len(tuples) == 2
}

func GetFullStudentName(studentName string, parentPhone string) string {
	return studentName + "/" + parentPhone
}
