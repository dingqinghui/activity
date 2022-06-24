/**
 * @Author: dingqinghui
 * @Description:
 * @File:  tools
 * @Version: 1.0.0
 * @Date: 2022/6/1 9:59
 */

package activity

import (
	"time"
)

var dateTimeFormat = "2006-01-02 15:04:05"

//
// nowTimestamp
// @Description: 获取当前时间戳(s)
// @return int64
//
func nowTimestamp() int64 {

	return time.Now().Unix()
}

//
// diffDayNum
// @Description: 比较时间间隔多少天
// @param now
// @param old
// @param hour 跨天小时
// @param timezone 时区
// @return int 间隔天数
//
func diffDayNum(now, old int64) int {
	hour := everydayUpdateHour
	tz := timeZero

	now += int64((tz - hour) * 3600)
	old += int64((tz - hour) * 3600)
	return int((now / 86400) - (old / 86400))
}

func timeStamp2Str(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(dateTimeFormat)
}

func isDifferDay(now, old int64) bool {
	return diffDayNum(now, old) > 0
}
