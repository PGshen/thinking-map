package utils

import (
	"time"
)

const (
	// TimeFormat 标准时间格式
	TimeFormat = "2006-01-02 15:04:05"
	// DateFormat 标准日期格式
	DateFormat = "2006-01-02"
	// TimeFormatWithZone 带时区的时间格式
	TimeFormatWithZone = "2006-01-02 15:04:05 -0700"
	// TimeFormatWithMilli 带毫秒的时间格式
	TimeFormatWithMilli = "2006-01-02 15:04:05.000"
)

// FormatTime 格式化时间为字符串
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

// FormatDate 格式化日期为字符串
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

// FormatTimeWithZone 格式化时间为带时区的字符串
func FormatTimeWithZone(t time.Time) string {
	return t.Format(TimeFormatWithZone)
}

// FormatTimeWithMilli 格式化时间为带毫秒的字符串
func FormatTimeWithMilli(t time.Time) string {
	return t.Format(TimeFormatWithMilli)
}

// ParseTime 解析时间字符串
func ParseTime(s string) (time.Time, error) {
	return time.Parse(TimeFormat, s)
}

// ParseDate 解析日期字符串
func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateFormat, s)
}

// ParseTimeWithZone 解析带时区的时间字符串
func ParseTimeWithZone(s string) (time.Time, error) {
	return time.Parse(TimeFormatWithZone, s)
}

// ParseTimeWithMilli 解析带毫秒的时间字符串
func ParseTimeWithMilli(s string) (time.Time, error) {
	return time.Parse(TimeFormatWithMilli, s)
}

// GetStartOfDay 获取一天的开始时间
func GetStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetEndOfDay 获取一天的结束时间
func GetEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// GetStartOfWeek 获取一周的开始时间
func GetStartOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return GetStartOfDay(t.AddDate(0, 0, -int(weekday-1)))
}

// GetEndOfWeek 获取一周的结束时间
func GetEndOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return GetEndOfDay(t.AddDate(0, 0, 7-int(weekday)))
}

// GetStartOfMonth 获取一个月的开始时间
func GetStartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// GetEndOfMonth 获取一个月的结束时间
func GetEndOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 999999999, t.Location())
}

// IsToday 判断是否是今天
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

// IsThisWeek 判断是否是本周
func IsThisWeek(t time.Time) bool {
	now := time.Now()
	startOfWeek := GetStartOfWeek(now)
	endOfWeek := GetEndOfWeek(now)
	return t.After(startOfWeek) && t.Before(endOfWeek)
}

// IsThisMonth 判断是否是本月
func IsThisMonth(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month()
}
