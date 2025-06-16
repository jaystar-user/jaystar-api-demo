package utils

import "time"

var (
	defaultLoc *time.Location
)

func init() {
	var err error
	defaultLoc, err = time.LoadLocation("Asia/Taipei")
	if err != nil {
		defaultLoc = time.UTC
	}
}

func GetLocation() *time.Location {
	return defaultLoc
}
