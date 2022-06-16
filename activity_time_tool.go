/**
 * @Author: dingqinghui
 * @Description:活动时间
 * @File:  player_activity_time_tool
 * @Version: 1.0.0
 * @Date: 2022/6/8 9:41
 */

package activity

import (
	"github.com/dingqinghui/activity/pb"
	"go.uber.org/zap"
)

type IActivityTime interface {
	getStartTime() int64
	getPredictionTime() int64
	getCloseTime() int64
	getEndTime() int64
}

type activityTimeBase struct {
	*pb.OperateActivity
	registerTime int64
	areaId       int32
}

func NewActivityTime(activity *pb.OperateActivity, registerTime int64, areaId int32) IActivityTime {
	base := &activityTimeBase{activity, registerTime, areaId}
	switch activity.GetTimeType() {
	case pb.OperateActivityTimeType_OPEN_SERVER_TIME:
		return &activityTimeOpenServer{base}
	case pb.OperateActivityTimeType_REGISTER_TIME:
		return &activityTimeRegister{base}
	case pb.OperateActivityTimeType_ABSOLUTE_TIME:
		return &activityTimeAbs{base}
	default:
		logWarn("invalid activity time type", zap.String("type", activity.GetTimeType().String()))
	}
	return nil
}

//
// activityTimeAbs
// @Description: 绝对时间处理
//
type activityTimeAbs struct {
	*activityTimeBase
}

func (m *activityTimeAbs) getPredictionTime() int64 {
	return m.GetPredictionTime()
}

func (m *activityTimeAbs) getStartTime() int64 {
	return m.GetStartTime()
}

func (m *activityTimeAbs) getCloseTime() int64 {
	return m.GetCloseDuration()
}

func (m *activityTimeAbs) getEndTime() int64 {
	return m.GetEndTime()
}

//
// activityTimeRegister
// @Description: 注册时间处理
//
type activityTimeRegister struct {
	*activityTimeBase
}

func (m *activityTimeRegister) getPredictionTime() int64 {
	return m.GetPredictionTime() + m.registerTime
}

func (m *activityTimeRegister) getStartTime() int64 {
	return m.GetStartTime() + m.registerTime
}

func (m *activityTimeRegister) getCloseTime() int64 {
	return m.GetCloseDuration() + m.registerTime
}

func (m *activityTimeRegister) getEndTime() int64 {
	return m.GetEndTime() + m.registerTime
}

//
// activityTimeOpenServer
// @Description: 开服时间处理
//
type activityTimeOpenServer struct {
	*activityTimeBase
}

func (m *activityTimeOpenServer) getPredictionTime() int64 {
	return m.GetPredictionTime() + m.getOpenServerTime()
}

func (m *activityTimeOpenServer) getStartTime() int64 {
	return m.GetStartTime() + m.getOpenServerTime()
}

func (m *activityTimeOpenServer) getCloseTime() int64 {
	return m.GetCloseDuration() + m.getOpenServerTime()
}

func (m *activityTimeOpenServer) getEndTime() int64 {
	return m.GetEndTime() + m.getOpenServerTime()
}

func (m *activityTimeOpenServer) getOpenServerTime() int64 {
	return GetAreaRegisterTime(m.areaId)
}
