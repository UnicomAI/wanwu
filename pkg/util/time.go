package util

import (
	"fmt"
	"time"
)

var UTC8 *time.Location

const timeMsFormat = "2006-01-02 15:04:05.000"

const timeFormat = "2006-01-02 15:04:05"
const dateFormat = "2006-01-02"

func InitTimeLocal() error {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return err
	}
	UTC8 = loc
	return nil
}

// WeekStartTime 某周的开始时间(周一0点)，-1上周，0本周，1下周
func WeekStartTime(t time.Time, week int) time.Time {
	offset := time.Monday - t.Weekday()
	if offset > 0 { // 周日特殊处理
		offset = -6
	}
	y, m, d := t.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, t.Location())
	return today.AddDate(0, 0, int(offset)+week*7)
}

func Time2Str(millSec int64) string {
	if millSec == 0 {
		return ""
	}
	return time.UnixMilli(millSec).In(UTC8).Format(timeFormat)
}

func Time2MsStr(millSec int64) string {
	if millSec == 0 {
		return ""
	}
	return time.UnixMilli(millSec).In(UTC8).Format(timeMsFormat)
}

func Str2Time(timeStr string) (int64, error) {
	t, err := time.ParseInLocation(timeFormat, timeStr, UTC8)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}

func Str2Date(timeStr string) (int64, error) {
	t, err := time.ParseInLocation(dateFormat, timeStr, UTC8)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}

func Time2Date(ts int64) string {
	return time.UnixMilli(ts).In(UTC8).Format(dateFormat) // 输出示例: 2025-05-09
}

func Date2Time(date string) (int64, error) {
	t, err := time.ParseInLocation(dateFormat, date, UTC8)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}

// DateRange 返回[startTs, endTs]闭区间日期列表
func DateRange(startTs, endTs int64) []string {
	if startTs > endTs {
		return nil
	}
	endDate := Time2Date(endTs)
	var ret []string
	for {
		date := Time2Date(startTs)
		ret = append(ret, date)
		if date == endDate {
			break
		}
		startTs = startTs + time.Hour.Milliseconds()*24
	}
	return ret
}

// 返回上一个周期和当前周期闭区间日期列表
func PreviousDateRange(startDate, endDate string) ([]string, []string, error) {
	// 1. 解析输入日期
	startAt, err := Date2Time(startDate)
	if err != nil {
		return nil, nil, err
	}
	endAt, err := Date2Time(endDate)
	if err != nil {
		return nil, nil, err
	}
	if startAt > endAt {
		return nil, nil, fmt.Errorf("startDate %v greater than endDate %v", startDate, endDate)
	}
	// 2. 计算上一周期时间戳区间（前后区间日期无重叠）
	deltaDura := endAt - startAt + 24*time.Hour.Milliseconds()
	pervStartTs := startAt - deltaDura
	pervEndTs := endAt - deltaDura
	// 3. 计算上一个周期，当前周期
	return DateRange(pervStartTs, pervEndTs), DateRange(startAt, endAt), nil
}

// NowSpanToHMS 将毫秒时间间隔转换为时分秒格式
// 输入：开始时间（int64类型）
// 输出：格式化字符串（如：1天02:15:30 或 02:15:30）
func NowSpanToHMS(startTime int64) string {
	milliseconds := time.Now().UnixMilli() - startTime
	if milliseconds < 1 {
		return "1ms"
	}
	if milliseconds < 1000 {
		return fmt.Sprintf("%dms", milliseconds)
	}

	// 计算总秒数
	totalSeconds := milliseconds / 1000

	// 计算天数
	days := totalSeconds / (24 * 3600)
	remainingSeconds := totalSeconds % (24 * 3600)

	// 计算小时、分钟、秒
	hours := remainingSeconds / 3600
	remainingSeconds %= 3600
	minutes := remainingSeconds / 60
	seconds := remainingSeconds % 60

	// 毫秒部分
	millis := milliseconds % 1000

	// 根据天数决定输出格式
	if days > 0 {
		return fmt.Sprintf("%dd%dh%dm%ds", days, hours, minutes, seconds)
	}

	// 如果有小时
	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	}

	// 如果只有分钟和秒
	if minutes > 0 {
		if seconds > 0 {
			return fmt.Sprintf("%dm%ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	}

	millis = millis / 100
	if millis == 0 {
		return fmt.Sprintf("%ds", seconds)
	}
	// 如果只有秒和毫秒
	return fmt.Sprintf("%d.%ds", seconds, millis)
}
