/**
 * @Author: dingqinghui
 * @Description:全局API接口
 * @File:  api
 * @Version: 1.0.0
 * @Date: 2022/6/1 10:06
 */

package global

import (
	"github.com/dingqinghui/activity/pb"
	"go.uber.org/zap"
)

//
// SetTimeZero
// @Description: 设置时区,默认东八区
// @param tz
//
func SetTimeZero(tz int) {
	timeZero = tz
}

//
// GetTimeZero
// @Description: 获取时区
// @return int
//
func GetTimeZero() int {
	return timeZero
}

//
// SetEverydayUpdateHour
// @Description:设置每日更新时间(每日几点算跨天)
// @param hour
//
func SetEverydayUpdateHour(hour int) {
	everydayUpdateHour = hour
}

//
// GetEverydayUpdateHour
// @Description: 获取每日更新小时
// @return int
//
func GetEverydayUpdateHour() int {
	return everydayUpdateHour
}

//
// GetAreaRegisterTime
// @Description: 获取区服注册时间
// @param areaId
// @return int64
//
func GetAreaRegisterTime(areaId int32) int64 {
	if areaRegisterTimeCb == nil {
		return NowTimestamp()
	}
	return areaRegisterTimeCb(areaId)
}

//
// Init
// @Description: 初始化全局管理器
// @param initData 全局活动数据
// @param dataCallback 全局活动数据更改回调函数
// @param artCb 获取区服开服时间函数
// @param l  日志处理器
//
func Init(initData []*pb.OperateActivity, dataCallback DataCmdFun, artCb AreaRegisterTimeFun, l *zap.Logger) {
	getGlobalOperateActivityMgr().init(initData, dataCallback)

	Logger = l

	areaRegisterTimeCb = artCb
}

//
// RangeAll
// @Description: 遍历所有活动
// @param f
//
func RangeAll(f func(*pb.OperateActivity)) {
	getGlobalOperateActivityMgr().rangeAll(f)
}

//
// Delete
// @Description: 删除活动，Gm撤回/删除时调用
// @param activityId
//
func Delete(activityId int64) {
	getGlobalOperateActivityMgr().delete(activityId)
	// getTimeTask().delete(activityId)
}

//
// Add
// @Description: 添加活动到全局管理器
// @param activity
//
func Add(activity *pb.OperateActivity) {
	if activity == nil {
		LogError("activity is nil")
		return
	}
	getGlobalOperateActivityMgr().addCache(activity)
	return
}

//
// GetActivity
// @Description: 获取活动
// @param activityId
// @return *pb.OperateActivity
//
func GetActivity(activityId int64) *pb.OperateActivity {
	return getGlobalOperateActivityMgr().getActivity(activityId)
}
