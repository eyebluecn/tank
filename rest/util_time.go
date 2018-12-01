package rest

import (
	"fmt"
	"time"
)

//将一个时间字符串转换成时间对象(yyyy-MM-dd HH:mm:ss)
func ConvertDateTimeStringToTime(timeString string) time.Time {
	local, _ := time.LoadLocation("Local")
	t, err := time.ParseInLocation("2006-01-02 15:04:05", timeString, local)
	if err != nil {
		panic(fmt.Sprintf("不能将%s转为时间类型", timeString))
	}
	return t
}

//将一个时间字符串转换成日期时间对象(yyyy-MM-dd HH:mm:ss)
func ConvertTimeToDateTimeString(time time.Time) string {
	return time.Local().Format("2006-01-02 15:04:05")
}

//将一个时间字符串转换成日期时间对象(yyyy-MM-dd HH:mm:ss)
func ConvertTimeToTimeString(time time.Time) string {
	return time.Local().Format("15:04:05")
}

//将一个时间字符串转换成日期对象(yyyy-MM-dd)
func ConvertTimeToDateString(time time.Time) string {
	return time.Local().Format("2006-01-02")
}

//一天中的最后一秒钟
func LastSecondOfDay(day time.Time) time.Time {
	local, _ := time.LoadLocation("Local")
	return time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, local)
}

//一天中的第一秒钟
func FirstSecondOfDay(day time.Time) time.Time {
	local, _ := time.LoadLocation("Local")
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, local)
}

//一天中的第一分钟
func FirstMinuteOfDay(day time.Time) time.Time {
	local, _ := time.LoadLocation("Local")
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 1, 0, 0, local)
}

//明天此刻的时间
func Tomorrow() time.Time {
	tomorrow := time.Now()
	tomorrow = tomorrow.AddDate(0, 0, 1)
	return tomorrow
}

//昨天此刻的时间
func Yesterday() time.Time {
	tomorrow := time.Now()
	tomorrow = tomorrow.AddDate(0, 0, -1)
	return tomorrow
}
