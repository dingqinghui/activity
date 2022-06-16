/**
 * @Author: dingqinghui
 * @Description:
 * @File:  tools
 * @Version: 1.0.0
 * @Date: 2022/6/1 9:59
 */

package global

import (
	"time"
)

var dateTimeFormat = "2006-01-02 15:04:05"

//
// NowTimestamp
// @Description: 获取当前时间戳(s)
// @return int64
//
func NowTimestamp() int64 {
	curTime := time.Now()
	nowTick := curTime.UnixNano() / 1e6
	return nowTick / 1e3
}

//
// DiffDayNum
// @Description: 比较时间间隔多少天
// @param now
// @param old
// @param hour 跨天小时
// @param timezone 时区
// @return int 间隔天数
//
func DiffDayNum(now, old int64) int {
	hour := everydayUpdateHour
	tz := timeZero

	now += int64((tz - hour) * 3600)
	old += int64((tz - hour) * 3600)
	return int((now / 86400) - (old / 86400))
}

func TimeStamp2Str(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(dateTimeFormat)
}

func IsDifferDay(now, old int64) bool {
	return DiffDayNum(now, old) > 0
}
