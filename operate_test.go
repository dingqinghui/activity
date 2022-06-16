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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
	"time"
)

//
// GlobalActivityDataUpdate
// @Description: 全局数据更改回调函数
// @param activity
// @param cmd
//
func GlobalActivityDataUpdate(activity *pb.OperateActivity, cmd global.DataCmd) {
	switch cmd {
	case global.DataAdd:

	case global.DataDelete:
		// 活动结束,删除db数据
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
func TestGlobal(t *testing.T) {
	// 全局数据初始化
	global.Init(nil, GlobalActivityDataUpdate, GetAreaStartTime, nil)
	global.Init(nil, GlobalActivityDataUpdate, GetAreaStartTime, global.WithLogger(zap.New(zapcore.NewTee())))
	global.Init(nil, GlobalActivityDataUpdate, GetAreaStartTime, global.WithLogConfig("", zap.DebugLevel))

	// 添加活动
	global.Add(&pb.OperateActivity{})
	// 删除活动
	global.Delete(1)
	// 设置时区 默认东八区
	global.SetTimeZero(8)
	// 设置每日更新时间(每日几点算跨天)
	global.SetEverydayUpdateHour(8)
}

// /////////////////////////////////////////////////////////////////玩家//////////////////////////////////////////////////////////////////////////

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

//
// PlayerActivityDataUpdate
// @Description: 玩家数据更改回调函数
// @param playerId
// @param activity
// @param cmd
//
func PlayerActivityDataUpdate(playerId int32, activityId int64, cmd player2.DataCmd, updateInfo *pb.OperateActivityDB) {
	switch cmd {
	case player2.DataAdd, player2.DataUpdate:
		// 更新玩家db数据
	case player2.DataDelete:
		// 删除玩家db数据
	default:
	}
}
func TestPlayer(t *testing.T) {
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
