/**
 * @Author: dingqinghui
 * @Description:
 * @File:  operate_test
 * @Version: 1.0.0
 * @Date: 2022/6/8 17:36
 */

package activity

import (
	"github.com/dingqinghui/activity/global"
	"github.com/dingqinghui/activity/pb"
	player2 "github.com/dingqinghui/activity/player"
	"strconv"
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	// 全局数据初始化
	global.Init(nil, GlobalActivityDataUpdate, GetAreaStartTime, nil)

	p := newPlayer()
	// 玩家登录
	_ = p.GetOperate().Login()
	// 签到
	_ = p.GetOperate().Sign(1, 1)

	// 其他接口

	// 触发任务
	p.GetOperate().TriggerCondition(func(conf *pb.Condition, taskInfo *pb.OperateTaskInfo) bool {
		if taskInfo.GetTaskState() != pb.OperateTaskState_OTS_Doing {
			return false
		}
		// 触发
		// 返回true 触发成功进行数据存档
		return true
	})
}

// /////////////////////////////////////////////////////////////////DB//////////////////////////////////////////////////////////////////////////

func LoadOneGlobalOperatorActivityData(activityId int64) *pb.OperateActivity {
	return nil
}

func DeleteGlobalOperatorActivityData(activityId int64) error {
	return nil
}

// /////////////////////////////////////////////////////////////////DB//////////////////////////////////////////////////////////////////////////

// /////////////////////////////////////////////////////////////////redis 通知//////////////////////////////////////////////////////////////////////////
//
// OnNotifyNewOperateActivity
// @Description: gm 添加活动通知
// @param _
// @param field
//
func OnNotifyNewOperateActivity(_ int32, field string) {
	activityId, _ := strconv.ParseInt(field, 10, 64)
	activity := LoadOneGlobalOperatorActivityData(activityId)
	if activity == nil {
		return
	}
	global.Add(activity)
}

//
// OnNotifyDeleteOperateActivity
// @Description: gm 撤回活动通知
// @param _
// @param field
//
func OnNotifyDeleteOperateActivity(_ int32, field string) {
	activityId, _ := strconv.ParseInt(field, 10, 64)
	if err := DeleteGlobalOperatorActivityData(activityId); err != nil {
		return
	}
	global.Delete(activityId)
	// 广播所有在线玩家
	//BroadcastInnerOnlinePlayer(GAME_CMD_INNER_DELETE_OPERATOR_ACTIVITY, activityId)
}

// /////////////////////////////////////////////////////////////////redis 通知//////////////////////////////////////////////////////////////////////////

// /////////////////////////////////////////////////////////////////回调//////////////////////////////////////////////////////////////////////////

//
// GlobalActivityDataUpdate
// @Description: 全局数据更改回调函数
// @param activity
// @param cmd
//
func GlobalActivityDataUpdate(activity *pb.OperateActivity, cmd global.DataCmd) {
	switch cmd {
	case global.DataAdd:
		// 活动开始,广播在线玩家
	case global.DataDelete:
		// 活动结束,删除db数据
	default:
	}
}

//
// PlayerActivityDataUpdate
// @Description: 玩家数据更改回调函数
// @param playerId
// @param activity
// @param cmd
//
func PlayerActivityDataUpdate(playerId int32, activityId int64, cmd global.DataCmd, updateInfo *pb.OperateActivityDB) {
	switch cmd {
	case global.DataAdd, global.DataUpdate:
		// 更新玩家db数据
	case global.DataDelete:
		// 删除玩家db数据
	default:
	}
}

//
// GetAreaStartTime
// @Description: 获取区服开服时间
// @param int32
// @return int64
//
func GetAreaStartTime(int32) int64 {
	return global.NowTimestamp()
}

// /////////////////////////////////////////////////////////////////回调//////////////////////////////////////////////////////////////////////////

// /////////////////////////////////////////////////////////////////内部消息//////////////////////////////////////////////////////////////////////////

//
// HandleNewOperatorActivity
// @Description:内部消息：添加活动
// @param player
// @param _
// @param msg
// @return interface{}
//
func HandleNewOperatorActivity(p *player, activity *pb.OperateActivity) interface{} {
	operator := p.GetOperate()
	if !operator.Add(activity) {
		return nil
	}

	// 通知客户端
	return nil
}

//
// HandleDeleteOperatorActivity
// @Description: 内部消息：删除活动
// @param player
// @param _
// @param msg
// @return interface{}
//
func HandleDeleteOperatorActivity(p *player, activity *pb.OperateActivity) interface{} {
	operator := p.GetOperate()
	operator.Delete(activity.GetId())

	// 通知客户端
	return nil
}

// /////////////////////////////////////////////////////////////////内部消息//////////////////////////////////////////////////////////////////////////

func newPlayer() *player {
	p := &player{}
	p.tick()
	return p
}

type player struct {
	operate *player2.ActivityMgr
}

func (p *player) GetId() int32 {
	return 1
}
func (p *player) OperateCheckCost(items []*pb.ItemData) error {
	// 自定义道具检测
	return nil
}
func (p *player) OperateAddReward(items []*pb.ItemData) error {
	// 自定义添加奖励
	return nil
}
func (p *player) OperateSubCost(items []*pb.ItemData) error {
	// 自定义扣除消耗
	return nil
}

func (p *player) OperateSendMail(items []*pb.ItemData) error {
	return nil
}

//
// tick
// @Description: 定时检测全局任务能够添加到玩家身上,和删除过期活动
// @receiver p
//
func (p *player) tick() {
	ticker := time.NewTicker(1)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.GetOperate().CheckNewAndDelete()
		}
	}
}

func (p *player) GetOperate() *player2.ActivityMgr {
	if p.operate == nil {
		// 创建玩家运营活动模块
		p.operate = player2.NewActivityMgr(p, 101, 10001, global.NowTimestamp(),
			PlayerActivityDataUpdate, nil)
	}
	return p.operate
}
