/**
 * @Author: dingqinghui
 * @Description:玩家活动管理模块
 * @File:  model_player
 * @Version: 1.0.0
 * @Date: 2022/5/31 17:33
 */

package activity

import (
	"errors"
	"github.com/dingqinghui/activity/pb"
	"go.uber.org/zap"
)

// 错误定义
var (
	// activityNotExist 活动不存在
	activityNotExist = errors.New("activity not exist")
	// templateNotExist 模板不存在
	templateNotExist = errors.New("template not exist")
)

// PlayerDataCmdFun 活动数据操作回调函数，当cmd == DataAdd时，updateInfo为活动完整DB数据，当cmd == DataUpdate，updateInfo为活动更改数据,未更改的数据赋值为nil
type PlayerDataCmdFun func(playerId int32, activityId int64, cmd DataCmd, updateInfo *pb.OperateActivityDB)

//
// IPlayer
// @Description: 玩家接口定义
//
type IPlayer interface {
	GetId() int32
	OperateCheckCost(activityId int64, cost []*pb.ItemData) error
	OperateAddReward(activityId int64, items []*pb.ItemData) error
	OperateSubCost(activityId int64, cost []*pb.ItemData) error
	OperateSendMail(activityId int64, items []*pb.ItemData) error
}

//
// NewPlayerActivityMgr
// @Description: 创建玩家活动管理器
// @param player 玩家对象
// @param areaId 玩家所属区服
// @param channel 玩家所属渠道
// @param registerTime 玩家注册时间
// @param changeDataCallback 玩家数据更改回调
// @param initData
// @return *PlayerActivityMgr
//
func NewPlayerActivityMgr(player IPlayer, areaId int32, channel int32, registerTime int64,
	changeDataCallback PlayerDataCmdFun) *PlayerActivityMgr {
	if player == nil {
		panic("operate player is nil")
	}
	m := &PlayerActivityMgr{
		player:              player,
		channel:             channel,
		registerTime:        registerTime,
		areaId:              areaId,
		changStatusCallback: changeDataCallback,
		activityMap:         make(map[int64]*Activity),
	}

	return m
}

//
// PlayerActivityMgr
// @Description: 玩家活动管理
//
type PlayerActivityMgr struct {
	//
	// player
	// @Description: 玩家
	//
	player IPlayer
	//
	// areaId
	// @Description: 玩家所属区服
	//
	areaId int32
	//
	// channel
	// @Description: 玩家所属渠道
	//
	channel int32
	//
	// registerTime
	// @Description: 玩家注册时间
	//
	registerTime int64
	//
	// activityMap
	// @Description: 所有活动实例
	//
	activityMap map[int64]*Activity
	//
	// changStatusCallback
	// @Description: 状态变化回调函数
	//
	changStatusCallback PlayerDataCmdFun
}

func (m *PlayerActivityMgr) InitData(initData map[int64]*pb.OperateActivityDB) {
	m.init(initData)
}

func (m *PlayerActivityMgr) getPlayer() IPlayer {
	return m.player
}

func (m *PlayerActivityMgr) getPlayerId() int32 {
	return m.player.GetId()
}
func (m *PlayerActivityMgr) getArea() int32 {
	return m.areaId
}
func (m *PlayerActivityMgr) getChannel() int32 {
	return m.channel
}
func (m *PlayerActivityMgr) getRegisterTime() int64 {
	return m.registerTime
}

func (m *PlayerActivityMgr) callActivityDataCmdFun(activityId int64, updateInfo *pb.OperateActivityDB, cmd DataCmd) {
	if m.changStatusCallback == nil {
		return
	}
	m.changStatusCallback(m.getPlayerId(), activityId, cmd, updateInfo)
	logInfo("回调活动数据操作", zap.Int32("playerId", m.getPlayerId()), zap.Any("cmd", cmd), zap.Int64("activityId", activityId), zap.Any("updateInfo", updateInfo))
}

func (m *PlayerActivityMgr) init(initData map[int64]*pb.OperateActivityDB) {
	// 分离过期活动和正常活动
	m.initActivity(initData)

	// 检测全局活动是否可添加
	m.CheckNewAndDelete()
}

//
// dbBatchDelete
// @Description: 批量db删除活动
// @receiver m
// @param deleteList
// @return error
//
func (m *PlayerActivityMgr) dbBatchDelete(deleteList []*pb.OperateActivityDB) error {
	if len(deleteList) <= 0 {
		return nil
	}
	for _, data := range deleteList {

		activity, _ := newActivity(data, m)
		if activity == nil {
			continue
		}

		_ = m.getPlayer().OperateSendMail(activity.getId(), activity.getCanReceiveReward(m.getPlayer()))

		m.callActivityDataCmdFun(data.GetActivityId(), nil, DataDelete)
	}
	return nil
}

//
// dbBatchAdd
// @Description: 批量db添加活动
// @receiver m
// @param addList
// @return error
//
func (m *PlayerActivityMgr) dbBatchAdd(addList []*pb.OperateActivityDB) error {
	if len(addList) <= 0 {
		return nil
	}

	for _, activity := range addList {
		m.callActivityDataCmdFun(activity.GetActivityId(), activity, DataAdd)
	}
	return nil
}

//
// initActivity
// @Description: 初始化活动实例，过滤掉无效任务，撤回和过期
// @receiver m
// @param initData
//
func (m *PlayerActivityMgr) initActivity(initData map[int64]*pb.OperateActivityDB) {
	addList := make([]*pb.OperateActivityDB, 0, 0)
	for _, activity := range initData {
		addList = append(addList, activity)
	}
	// 批量添加任务活动缓存
	m.addActivityList(addList)
}

func (m *PlayerActivityMgr) generateActivityCommonData(conf *pb.OperateActivity) *pb.OperateActivityDB {
	dbData := &pb.OperateActivityDB{
		ActivityId:   conf.GetId(),
		ActivityList: make(map[int32]*pb.ActivityDBList),
		GotScores:    make(map[int32]bool),
	}

	for _, group := range conf.GetPreConditionGroup() {
		groupData := new(pb.TaskGroup)
		dbData.PreTaskGroup = append(dbData.PreTaskGroup, groupData)
		for range group.GetPreCondition() {
			groupData.PreTaskInfos = append(groupData.PreTaskInfos, new(pb.OperateTaskInfo))
		}
	}
	return dbData
}

//
// checkAndAddGlobalActivity
// @Description: 检测全局活动是否可添加到玩家
// @receiver m
//
func (m *PlayerActivityMgr) checkAndAddGlobalActivity() {
	var addCacheList []*pb.OperateActivityDB
	RangeAll(func(conf *pb.OperateActivity) {
		if m.getActivity(conf.GetId()) != nil {
			return
		}
		if m.checkAddCondition(conf) {
			dbData := m.generateActivityCommonData(conf)
			addCacheList = append(addCacheList, dbData)
		}
	})

	// 批量添加任务活动缓存
	m.addActivityList(addCacheList)

	// 批量db添加活动
	_ = m.dbBatchAdd(addCacheList)
	return
}

//
// Add
// @Description: 添加运营活动
// @receiver m
// @param activity
// @return bool true:添加成功 false:条件不满足
//
func (m *PlayerActivityMgr) Add(conf *pb.OperateActivity) bool {
	if !m.checkAddCondition(conf) {
		return false
	}

	dbData := m.generateActivityCommonData(conf)

	// 添加实例
	m.addActivityList([]*pb.OperateActivityDB{dbData})

	// 回调通知添加成功
	_ = m.dbBatchAdd([]*pb.OperateActivityDB{dbData})
	return true
}

//
// CheckNewAndDelete
// @Description: 检测添加新活动和删除旧活动
// @receiver m
//
func (m *PlayerActivityMgr) CheckNewAndDelete() {
	m.checkAndAddGlobalActivity()
	m.checkDeleteActivity()
}

func (m *PlayerActivityMgr) checkDeleteActivity() {
	m.rangeAll(func(activity *Activity) {
		// 未撤回&&未过期
		conf := GetActivity(activity.getId())
		if conf != nil && !activity.isExpire() {
			return
		}
		m.Delete(activity.getId())
	})
}

//
// checkExpire
// @Description: 活动是否过期
// @receiver m
// @param Activity
// @return bool true:过期
//
func (m *PlayerActivityMgr) checkExpire(conf *pb.OperateActivity) bool {
	timeTool := NewActivityTime(conf, m.getRegisterTime(), m.getArea())
	if timeTool == nil {
		return true
	}
	return nowTimestamp() >= timeTool.getCloseTime()
}

//
// checkAddCondition
// @Description: 检测添加条件是否满足
// @receiver m
// @param activity
// @return bool
//
func (m *PlayerActivityMgr) checkAddCondition(activity *pb.OperateActivity) bool {
	// 区服
	if !m.checkArea(activity) {
		return false
	}

	// 渠道不满足
	if !m.checkChannel(activity) {
		return false
	}

	// 检测时间
	timeTool := NewActivityTime(activity, m.getRegisterTime(), m.getArea())
	if timeTool == nil {
		return false
	}
	nowTime := nowTimestamp()
	if nowTime < timeTool.getPredictionTime() || nowTime > timeTool.getCloseTime() {
		return false
	}
	return true
}

//
// Delete
// @Description: 删除活动
// @receiver m
// @param activityId 活动Id
// @return bool
//
func (m *PlayerActivityMgr) Delete(activityId int64) bool {
	// 撤回直接删除活动
	conf := GetActivity(activityId)
	if conf == nil {
		m.callActivityDataCmdFun(activityId, nil, DataDelete)
		delete(m.activityMap, activityId)
		logInfo("配置不存在删除运营活动实例", zap.Int32("playerId", m.getPlayerId()), zap.Int64("activityId", activityId))
		return true
	}

	activity, ok := m.activityMap[activityId]
	if !ok {
		return false
	}
	_ = m.getPlayer().OperateSendMail(activityId, activity.getCanReceiveReward(m.getPlayer()))
	m.callActivityDataCmdFun(activityId, nil, DataDelete)
	delete(m.activityMap, activityId)
	logInfo("删除运营活动实例", zap.Int32("playerId", m.getPlayerId()), zap.Int64("activityId", activityId), zap.Any("Activity", activity))
	return true
}

//
// addActivityList
// @Description: 批量添加任务活动缓存
// @receiver m
// @param list
//
func (m *PlayerActivityMgr) addActivityList(list []*pb.OperateActivityDB) {
	for _, data := range list {
		if _, ok := m.activityMap[data.GetActivityId()]; ok {
			return
		}
		activity, err := newActivity(data, m)
		if err != nil {
			logError("activity conf is nil delete activity", zap.Error(err))
			m.callActivityDataCmdFun(data.GetActivityId(), nil, DataDelete)
			continue
		}
		m.activityMap[activity.getId()] = activity
		logInfo("添加运营活动实例成功", zap.Int32("playerId", m.getPlayerId()), zap.Int64("id", activity.getId()))
	}
	return
}

//
// rangeAll
// @Description: 遍历所有活动
// @receiver m
// @param f
//
func (m *PlayerActivityMgr) rangeAll(f func(act *Activity)) {
	if f == nil {
		return
	}
	for _, v := range m.activityMap {
		f(v)
	}
}

//
// RangeAllOpen
// @Description: 遍历所有进行中活动
// @receiver m
// @param f
//
func (m *PlayerActivityMgr) RangeAllOpen(f func(act *Activity)) {
	if f == nil {
		return
	}
	for _, v := range m.activityMap {
		if err := v.invalid(); err != nil {
			return
		}
		f(v)
	}
}

//
// getActivity
// @Description: 获取活动势力
// @receiver m
// @param activityId
// @return *Activity
//
func (m *PlayerActivityMgr) getActivity(activityId int64) *Activity {
	activity, ok := m.activityMap[activityId]
	if !ok {
		return nil
	}
	return activity
}

//
// getStartActivity
// @Description: 根据Id获取开启的活动
// @receiver m
// @param activityId
// @return *Activity
//
func (m *PlayerActivityMgr) getStartActivity(activityId int64) *Activity {
	activity, ok := m.activityMap[activityId]
	if !ok {
		return nil
	}
	if err := activity.invalid(); err != nil {
		logWarn("活动不可用", zap.Error(err), zap.Int32("playerId", m.getPlayerId()), zap.Int64("activityId", activityId))
		return nil
	}
	return m.activityMap[activityId]
}

//
// checkArea
// @Description: 检测区服是否满足
// @receiver m
// @param area
// @return bool true:满足
//
func (m *PlayerActivityMgr) checkArea(activity *pb.OperateActivity) bool {
	servers := activity.GetServers()
	if len(servers) <= 0 {
		return true
	}
	for _, server := range servers {
		if m.getArea() == server {
			return true
		}
	}
	return false
}

//
// checkChannel
// @Description: 检测区服是否满足
// @receiver m
// @param channel
// @return bool true:满足
//
func (m *PlayerActivityMgr) checkChannel(activity *pb.OperateActivity) bool {
	channels := activity.GetChannel()
	if len(channels) <= 0 {
		return true
	}
	for _, ch := range channels {
		if m.getChannel() == ch {
			return true
		}
	}
	return false
}

// ////////////////////////////////////////////////////////导出接口///////////////////////////////////////////////////////////////////////////////////////

//
// Login
// @Description: 被动触发,玩家登录
// @receiver m
// @param player
// @return error
//
func (m *PlayerActivityMgr) Login() error {
	m.RangeAllOpen(func(act *Activity) {
		act.rangeTemplates(func(template iTemplate) {
			if template.getType() != pb.ActivityTemplateType_SIGN_IN_TYPE {
				return
			}
			sign, ok := template.(*signTemplate)
			if !ok {
				return
			}
			// 登录主动触发
			if !sign.isLoginTrigger() {
				return
			}
			if err := sign.sign(m.getPlayer()); err != nil {
				return
			}
		})
	})
	return nil
}

//
// TriggerCondition
// @Description: 遍历玩家所有活动任务
// @receiver m
// @param f
//
func (m *PlayerActivityMgr) TriggerCondition(f func(conf *pb.Condition, taskInfo *pb.OperateTaskInfo) bool) {
	m.rangeAll(func(activity *Activity) {
		activity.rangeAllCondition(f)
	})
}

//
// Sign
// @Description: 签到
// @receiver m
// @param activityId 活动Id
// @param index	活动模板索引
// @return error
//
func (m *PlayerActivityMgr) Sign(activityId int64, index int) error {
	activity := m.getStartActivity(activityId)
	if activity == nil {
		return activityNotExist
	}
	template := activity.getSignTemplate(index)
	if template == nil {
		return templateNotExist
	}
	// 登录主动触发
	if template.isLoginTrigger() {
		return errors.New("sign trigger error") // 签到触发类型错误
	}
	if err := template.sign(m.getPlayer()); err != nil {
		return err
	}
	return nil
}

//
// SignGetReward
// @Description: 领取签到奖励
// @receiver m
// @param activityId 活动Id
// @param index 活动模板索引
// @param day 领取哪天
// @return error
//
func (m *PlayerActivityMgr) SignGetReward(activityId int64, index int, day int32) error {
	activity := m.getStartActivity(activityId)
	if activity == nil {
		return activityNotExist
	}
	template := activity.getSignTemplate(index)
	if template == nil {
		return templateNotExist
	}
	//// 自动发送签到奖励
	//if template.isAutoGetReward() {
	//	return errors.New("sign reward is auto send")
	//}
	if err := template.getReward(m.getPlayer(), day); err != nil {
		logError("领取签到奖励", zap.Int32("playerId", m.getPlayerId()), zap.Int64("activityId", activityId), zap.Int32("day", day), zap.Error(err))
		return err
	}
	// 存档
	template.saveDB()
	logError("领取签到奖励", zap.Int32("playerId", m.getPlayerId()), zap.Int64("activityId", activityId), zap.Int32("day", day))
	return nil
}

//
// SignRepair
// @Description: 补签
// @receiver m
// @param activityId 活动Id
// @param index 活动模板索引
// @return error
//
func (m *PlayerActivityMgr) SignRepair(activityId int64, index int) error {
	activity := m.getStartActivity(activityId)
	if activity == nil {
		return activityNotExist
	}
	template := activity.getSignTemplate(index)
	if template == nil {
		return templateNotExist
	}
	if err := template.repair(m.getPlayer()); err != nil {
		return err
	}
	return nil
}

//
// GetTaskReward
// @Description: 领取任务奖励
// @receiver m
// @param activityId 活动Id
// @param index 活动模板索引
// @param taskIndex 任务索引
// @return error
//
func (m *PlayerActivityMgr) GetTaskReward(activityId int64, index int, taskIndex int32) error {
	activity := m.getStartActivity(activityId)
	if activity == nil {
		return activityNotExist
	}
	template := activity.getTaskTemplate(index)
	if template == nil {
		return templateNotExist
	}
	if err := template.finishTask(m.getPlayer(), taskIndex); err != nil {
		return err
	}
	return nil
}

//
// ShopBuyGoods
// @Description: 购买商品
// @receiver m
// @param activityId 活动Id
// @param index 活动模板索引
// @param goodsIndex 商品索引
// @return error
//
func (m *PlayerActivityMgr) ShopBuyGoods(activityId int64, index int, goodsIndex int) error {
	activity := m.getStartActivity(activityId)
	if activity == nil {
		return activityNotExist
	}
	template := activity.getShopTemplate(index)
	if template == nil {
		return templateNotExist
	}
	if err := template.buy(m.getPlayer(), goodsIndex); err != nil {
		return err
	}
	return nil
}

//
// GetScoreReward
// @Description: 获取积分奖励
// @receiver m
// @param activityId 活动Id
// @param index 活动模板索引
// @return error
//
func (m *PlayerActivityMgr) GetScoreReward(activityId int64, index int) error {
	activity := m.getStartActivity(activityId)
	if activity == nil {
		return activityNotExist
	}
	if err := activity.getScoreReward(m.getPlayer(), index); err != nil {
		return err
	}
	return nil
}

//
// PackAllOpenActivity
// @Description: 打包所有开启活动
// @receiver m
// @return *OperateGetListS2C
//
func (m *PlayerActivityMgr) PackAllOpenActivity() *pb.OperateGetListS2C {
	s2c := &pb.OperateGetListS2C{}
	m.rangeAll(func(activity *Activity) {
		s2c.List = append(s2c.List, activity.getClientData())
	})
	return s2c
}

//
// PackOneActivity
// @Description: 打包一个活动所有数据
// @receiver m
// @param activityId
// @return *OperateNewS2C
//
func (m *PlayerActivityMgr) PackOneActivity(activityId int64) *pb.OperateNewS2C {
	activity := m.getActivity(activityId)
	if activity == nil {
		return nil
	}
	s2c := &pb.OperateNewS2C{}
	s2c.List = append(s2c.List, activity.getClientData())
	return s2c
}

//
// resetTaskByType
// @Description: 任务重置检测
// @receiver m
// @param resetType
//
func (m *PlayerActivityMgr) resetTaskByType(resetType pb.TaskRefreshType) {
	// 遍历所有任务
	m.rangeAll(func(activity *Activity) {
		var rewardList []*pb.ItemData
		f := func(conf *pb.Condition, taskInfo *pb.OperateTaskInfo) bool {
			if conf.GetRefreshType() != resetType {
				return false
			}
			// 添加未领取奖励
			if taskInfo.GetTaskState() != pb.OperateTaskState_OTS_Finish {
				rewardList = append(rewardList, conf.GetRewardList()...)
			}

			// 重置任务状态和进度
			*taskInfo = pb.OperateTaskInfo{}

			return true
		}
		activity.rangeAllCondition(f)

		// 邮件发送未领取奖励
		if len(rewardList) > 0 {
			_ = m.getPlayer().OperateSendMail(activity.getId(), rewardList)
		}
	})
}

//
// resetEveryDaySignRepairCount
// @Description: 重置补签次数
// @receiver m
//
func (m *PlayerActivityMgr) resetEveryDaySignRepairCount() {
	m.RangeAllOpen(func(act *Activity) {
		act.rangeTemplates(func(template iTemplate) {
			if template.getType() != pb.ActivityTemplateType_SIGN_IN_TYPE {
				return
			}
			sign, ok := template.(*signTemplate)
			if !ok {
				return
			}
			sign.resetEveryDayRepairCount()
		})
	})
}

//
// OnNewDay
// @Description: 跨天
// @receiver m
//
func (m *PlayerActivityMgr) OnNewDay() {
	// 重置任务状态
	m.resetTaskByType(pb.TaskRefreshType_TRT_DAY)
	// 重置签到每日补签次数
	m.resetEveryDaySignRepairCount()
	// 触发一次自动签到
	_ = m.Login()
}

//
// OnNewWeek
// @Description: 跨周
// @receiver m
//
func (m *PlayerActivityMgr) OnNewWeek() {
	// 重置任务状态
	m.resetTaskByType(pb.TaskRefreshType_TRT_WEEK)
}

//
// OnNewMonth
// @Description: 跨月
// @receiver m
//
func (m *PlayerActivityMgr) OnNewMonth() {
	// 重置任务状态
	m.resetTaskByType(pb.TaskRefreshType_TRT_MONTH)
}
