package util

import (
	"fmt"
	"time"
)

//convert time string(yyyy-MM-dd HH:mm:ss) to Time object
func ConvertDateTimeStringToTime(timeString string) time.Time {
	local, _ := time.LoadLocation("Local")
	t, err := time.ParseInLocation("2006-01-02 15:04:05", timeString, local)
	if err != nil {
		panic(fmt.Sprintf("cannot convert %s to Time", timeString))
	}
	return t
}

//convert Time object to string(yyyy-MM-dd HH:mm:ss)
func ConvertTimeToDateTimeString(time time.Time) string {
	return time.Local().Format("2006-01-02 15:04:05")
}

//convert Time object to string(HH:mm:ss)
func ConvertTimeToTimeString(time time.Time) string {
	return time.Local().Format("15:04:05")
}

//convert Time object to string(yyyy-MM-dd)
func ConvertTimeToDateString(time time.Time) string {
	return time.Local().Format("2006-01-02")
}

func LastSecondOfDay(day time.Time) time.Time {
	local, _ := time.LoadLocation("Local")
	return time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, local)
}

func FirstSecondOfDay(day time.Time) time.Time {
	local, _ := time.LoadLocation("Local")
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, local)
}

func FirstMinuteOfDay(day time.Time) time.Time {
	local, _ := time.LoadLocation("Local")
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 1, 0, 0, local)
}

//Tomorrow right now
func Tomorrow() time.Time {
	tomorrow := time.Now()
	tomorrow = tomorrow.AddDate(0, 0, 1)
	return tomorrow
}

//Yesterday right now
func Yesterday() time.Time {
	tomorrow := time.Now()
	tomorrow = tomorrow.AddDate(0, 0, -1)
	return tomorrow
}
