package utils

import (
	"strings"
	"time"
)

// 日期格式：模仿java中的结构体
type DateStyle string

const (
	YYYYMMDDHHMMSS = "yyyyMMddHHmmss"

	YYMMDDHHMMSS = "yyMMddHHmmss"

	YYMMDDHHMM = "yyMMddHHmm"

	YYMMDDHH = "yyMMddHH"

	YYMMDD = "yyMMdd"

	YYYY_MM_DD_HH_MM_SS_SSS = "yyyy-MM-dd HH:mm:ss.SSS"

	YYYY_MM_DD_HH_MM_SS_SSS_EN = "yyyy/MM/dd HH:mm:ss.SSS"

	YYYY_MM_DD_HH_MM_SS_CN = "yyyy年MM月dd日 HH:mm:ss"

	HH_MM_SS_MS = "HH:mm:ss.SSS"
)

// 日期转字符串
func FormatDate(date time.Time, dateStyle DateStyle) string {
	layout := string(dateStyle)
	layout = strings.Replace(layout, "yyyy", "2006", 1)
	layout = strings.Replace(layout, "yy", "06", 1)
	layout = strings.Replace(layout, "MM", "01", 1)
	layout = strings.Replace(layout, "dd", "02", 1)
	layout = strings.Replace(layout, "HH", "15", 1)
	layout = strings.Replace(layout, "mm", "04", 1)
	layout = strings.Replace(layout, "ss", "05", 1)
	layout = strings.Replace(layout, "SSS", "000", -1)
	return date.Format(layout)
}
