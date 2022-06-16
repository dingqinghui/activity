/**
 * @Author: dingqinghui
 * @Description:全局活动管理器
 * @File:  api
 * @Version: 1.0.0
 * @Date: 2022/6/1 10:06
 */

package activity

import (
	"github.com/dingqinghui/activity/pb"
	"go.uber.org/zap"
	"reflect"
	"sync"
)

type (
	// AreaRegisterTimeFun 获取区服注册时间
	AreaRegisterTimeFun func(int32) int64
)

var (
	// everydayUpdateHour 每日刷新时间
	everydayUpdateHour = 5
	// timeZero 时区
	timeZero = 8
	// 获取区服注册时间回调
	areaRegisterTimeCb AreaRegisterTimeFun
)

// 活动全局管理模块
var (
	globalOperateActivityMgr *operatorActivityMgr
	onceActivityMgr          sync.Once
)

//
// getGlobalOperateActivityMgr
// @Description: 单例
// @return *operatorActivityMgr
//
func getGlobalOperateActivityMgr() *operatorActivityMgr {
	onceActivityMgr.Do(func() {
		globalOperateActivityMgr = new(operatorActivityMgr)
	})
	return globalOperateActivityMgr
}

//
// operatorActivityMgr
// @Description: 全局活动列表管理器
//
type operatorActivityMgr struct {
	//
	// activityMap
	// @Description: 活动列表
	//
	activityMap sync.Map

	//
	// changStatusCallback
	// @Description: 状态变化回调函数
	//
	changStatusCallback DataCmdFun
}

func (m *operatorActivityMgr) init(initData []*pb.OperateActivity, cb DataCmdFun) {
	m.changStatusCallback = cb

	deleteList := make([]*pb.OperateActivity, 0, 0)
	for _, activity := range initData {
		if m.checkExpire(activity) {
			deleteList = append(deleteList, activity)
			continue
		}
		m.addCache(activity)
	}
	_ = m.batchDelete(deleteList)
}

//
// batchDelete
// @Description: 批量删除活动
// @receiver m
// @param deleteList
// @return error
//
func (m *operatorActivityMgr) batchDelete(deleteList []*pb.OperateActivity) error {
	if len(deleteList) == 0 {
		return nil
	}
	for _, activity := range deleteList {
		m.activityMap.Delete(activity.GetId())
		m.callDataCmdFun(activity, DataDelete)
		logInfo("db删除过期运营活动数据", zap.Int64("activityId", activity.GetId()))
	}
	return nil
}

func (m *operatorActivityMgr) delete(activityId int64) {
	m.activityMap.Delete(activityId)
	logInfo("删除运营活动数据", zap.Int64("activityId", activityId))
}

func (m *operatorActivityMgr) callDataCmdFun(activity *pb.OperateActivity, cmd DataCmd) {
	if m.changStatusCallback == nil {
		return
	}
	m.changStatusCallback(activity, cmd)
}

//
// addCache
// @Description:添加活动实例
// @receiver m
// @param pActivity 活动原始数据
// @return bool true:添加成功
//
func (m *operatorActivityMgr) addCache(pActivity *pb.OperateActivity) bool {
	activity := m.getActivity(pActivity.GetId())
	if activity != nil {
		logError("添加失败已经存在运营活动实例", zap.Int64("activityId", pActivity.GetId()))
		return false
	}
	m.callDataCmdFun(activity, DataAdd)
	m.activityMap.Store(pActivity.GetId(), pActivity)

	logInfo("添加运营活动实例成功", zap.Int64("activityId", pActivity.GetId()), zap.Any("activity", pActivity))
	return true
}

//
// rangeAll
// @Description: 遍历所有未过期的活动
// @receiver m
// @param f
//
func (m *operatorActivityMgr) rangeAll(f func(*pb.OperateActivity)) {
	if f == nil {
		return
	}
	deleteList := make([]*pb.OperateActivity, 0, 0)
	m.activityMap.Range(func(key, value interface{}) bool {
		activity, ok := value.(*pb.OperateActivity)
		if !ok {
			logError("invalid activity data type", zap.String("dataType", reflect.TypeOf(activity).String()))
			return true
		}
		if m.checkExpire(activity) {
			deleteList = append(deleteList, activity)
			return true
		}
		f(activity)
		return true
	})
	_ = m.batchDelete(deleteList)
}

//
// getActivity
// @Description: 获取未过期活动实例
// @receiver m
// @param activityId
// @return *OperateActivity
//
func (m *operatorActivityMgr) getActivity(activityId int64) *pb.OperateActivity {
	value, ok := m.activityMap.Load(activityId)
	if !ok {
		return nil
	}
	activity, ok := value.(*pb.OperateActivity)
	if !ok {
		logError("invalid activity data type", zap.String("dataType", reflect.TypeOf(activity).String()))
		return nil
	}
	deleteList := make([]*pb.OperateActivity, 0, 0)
	if m.checkExpire(activity) {
		deleteList = append(deleteList, activity)
		return nil
	}
	_ = m.batchDelete(deleteList)
	return activity
}

// checkExpire
// @Description: 检测全局活动是否过期,只能检测绝对时间
// @receiver m
// @param activity
// @return bool true:过期
//
func (m *operatorActivityMgr) checkExpire(activity *pb.OperateActivity) bool {
	if activity.GetTimeType() != pb.OperateActivityTimeType_ABSOLUTE_TIME {
		return false
	}
	return nowTimestamp() >= activity.GetEndTime()
}
