package timeUtil

import "time"

func GetLocalLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return time.UTC
	}

	return loc
}
