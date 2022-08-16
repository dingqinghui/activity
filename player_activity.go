/**
 * @Author: dingqinghui
 * @Description:玩家单个活动
 * @File:  Activity
 * @Version: 1.0.0
 * @Date: 2022/6/6 12:04
 */

package activity

import (
	"errors"
	"github.com/dingqinghui/activity/pb"
	"go.uber.org/zap"
)

// RangeTaskFunType 触发任务遍历函数  return:true 触发任务成功
type RangeTaskFunType func(conf *pb.Condition, taskInfo *pb.OperateTaskInfo) bool

type (
	// DataCmd 活动数据操作
	DataCmd int
	// DataCmdFun 活动数据操作回调函数
	DataCmdFun func(activity *pb.OperateActivity, cmd DataCmd)
)

// 数据操作指令
var (
	// DataAdd 添加活动
	DataAdd DataCmd = 1
	// DataDelete 删除活动
	DataDelete DataCmd = 2
	// DataUpdate 更新活动
	DataUpdate DataCmd = 3
)

//
// Activity
// @Description: 活动实例
//
type Activity struct {

	//
	// m
	// @Description: 所属管理器
	//
	mgr *PlayerActivityMgr

	//
	// templates
	// @Description: 活动模板处理器
	//
	templates map[int32][]iTemplate

	//
	// timeTool
	// @Description: 活动时间处理器
	//
	//timeTool IActivityTime

	//
	// dbData
	// @Description: 活动DB数据
	//
	dbData *pb.OperateActivityDB

	//
	// conf
	// @Description: 配置数据
	//
	conf *pb.OperateActivity
}

func newActivity(dbData *pb.OperateActivityDB, mgr *PlayerActivityMgr) (*Activity, error) {
	if mgr == nil {
		return nil, errors.New("mgr is nil")
	}
	if dbData == nil {
		return nil, errors.New("db data is nil")
	}
	conf := GetActivity(dbData.GetActivityId())
	if conf == nil {
		return nil, errors.New("conf is nil")
	}

	timeTool := NewActivityTime(conf, mgr.getRegisterTime(), mgr.getArea())

	// 拷贝global配置数据,转换相对时间为时间戳
	cConf := &pb.OperateActivity{}
	if err := deepCopy(conf, cConf); err != nil {
		return nil, err
	}
	cConf.PredictionTime = timeTool.getPredictionTime()
	cConf.StartTime = timeTool.getStartTime()
	cConf.EndTime = timeTool.getEndTime()
	cConf.CloseDuration = timeTool.getCloseTime()

	activity := &Activity{
		mgr:       mgr,
		templates: make(map[int32][]iTemplate),
		dbData:    dbData,
		conf:      conf,
		//timeTool:  timeTool,
	}
	activity.init()
	return activity, nil
}

func (m *Activity) init() {
	// 初始化所有模板
	for day, list := range m.getConf().GetActivityList() {
		for index, tplConf := range list.List {
			dbData := m.getTemplateData(day, index)
			isInit := dbData == nil

			t := m.addTemplate(day, int32(index), tplConf, dbData)
			if isInit {
				m.saveTemplateData(day, index, t.getDbData())
			}
		}
	}
}

func (m *Activity) saveTemplateData(day int32, index int, data *pb.ActivityTemplateDB) {
	templates, ok := m.getDbData().GetActivityList()[day]
	if !ok {
		m.getDbData().GetActivityList()[day] = &pb.ActivityDBList{
			List: make(map[int32]*pb.ActivityTemplateDB),
		}
		templates = m.dbData.GetActivityList()[day]
	}
	templates.List[int32(index)] = data
}

func (m *Activity) getTemplateData(day int32, index int) *pb.ActivityTemplateDB {
	templates := m.getDbData().GetActivityList()[day]
	if templates == nil {
		return nil
	}
	if index >= len(templates.GetList()) {
		return nil
	}
	return templates.GetList()[int32(index)]
}

//
// getId
// @Description: 获取活动Id
// @receiver m
// @return int64
//
func (m *Activity) getId() int64 {
	return m.getDbData().GetActivityId()
}

//
// getConf
// @Description: 获取GM后台配置数据
// @receiver m
// @return *OperateActivity
//
func (m *Activity) getConf() *pb.OperateActivity {
	return m.conf
}

//
// getDbData
// @Description: 获取活动数据
// @receiver m
// @return *OperateActivity
//
func (m *Activity) getDbData() *pb.OperateActivityDB {
	return m.dbData
}

//
// addTemplate
// @Description: 添加模板
// @receiver m
// @param day
// @param tpl
//
func (m *Activity) addTemplate(day int32, index int32, tplConf *pb.ActivityTemplate, dbData *pb.ActivityTemplateDB) iTemplate {
	t := newTemplate(day, index, tplConf, m, dbData)
	m.templates[day] = append(m.templates[day], t)
	logInfo("activity add template", zap.Int32("playerId", m.mgr.getPlayerId()), zap.Int64("activityId", m.getId()), zap.Any("tpl", tplConf))
	return t
}

//
// finishedPreCondition
// @Description: 前置任务是否完成
// @receiver m
// @return bool true:完成
//
func (m *Activity) finishedPreCondition() bool {
	defaultRet := true
	preConditionAllFinished := m.getConf().GetNeedPreCondAllFinished()
	for _, taskInfo := range m.getDbData().GetPreTaskInfos() {
		if preConditionAllFinished {
			// 全部完成
			if taskInfo.GetTaskState() == pb.OperateTaskState_OTS_Doing {
				return false
			}
			defaultRet = true
		} else {
			// 只完成一个
			if taskInfo.GetTaskState() != pb.OperateTaskState_OTS_Doing {
				return true
			}
			defaultRet = false
		}
	}
	return defaultRet
}

//
// invalid
// @Description: 活动是否无效
// @receiver m
// @return bool true:无效
//
func (m *Activity) invalid() error {
	// 未完成前置任务
	if !m.finishedPreCondition() {
		return errors.New("活动前置条件未满足")
	}
	if !m.isOpenTime() {
		return errors.New("活动未到开启时间")
	}
	return nil
}

//
// openDay
// @Description: 开启天数
// @receiver m
// @return int32
//
func (m *Activity) openDay() int32 {
	return int32(diffDayNum(nowTimestamp(), m.getConf().GetStartTime())) + 1
}

//
// isExpire
// @Description: 是否过期
// @receiver m
// @return bool true:过期
//
func (m *Activity) isExpire() bool {
	return nowTimestamp() >= m.getConf().GetCloseDuration()
}

//
// getTemplates
// @Description: 获取所有已开启模板
// @receiver m
// @return []iTemplate
//
func (m *Activity) getTemplates() []iTemplate {
	conf := m.getConf()
	// 时间嵌套返回当天开启的模板
	if conf.GetIsNestedActivity() {
		day := m.openDay()
		return m.templates[day]
	} else {
		list := make([]iTemplate, 0, 0)
		for _, v := range m.templates {
			list = append(list, v...)
		}
		return list
	}
}

//
// getTemplate
// @Description: 获取特定索引的模板
// @receiver m
// @param index
// @return iTemplate
//
func (m *Activity) getTemplate(index int) iTemplate {
	templates := m.getTemplates()
	if index >= len(templates) {
		return nil
	}
	return templates[index]
}

//
// rangeTemplates
// @Description: 遍历所有开启模板
// @receiver m
// @param f
//
func (m *Activity) rangeTemplates(f func(template iTemplate)) {
	if f == nil {
		return
	}
	list := m.getTemplates()
	for _, v := range list {
		f(v)
	}
}

//
// isOpen
// @Description: 活动是否进行中
// @receiver m
// @return bool
//
func (m *Activity) isOpenTime() bool {
	nowTime := nowTimestamp()
	return m.getConf().GetStartTime() <= nowTime && nowTime <= m.getConf().GetEndTime()
}

func (m *Activity) callUpdateStatusFun(updateInfo *pb.OperateActivityDB, status DataCmd) {
	m.mgr.callActivityDataCmdFun(m.getId(), updateInfo, status)
}

//
// rangeAllCondition
// @Description: 遍历活动所有任务
// @receiver m
// @param f
//
func (m *Activity) rangeAllCondition(f RangeTaskFunType) {
	conf := m.getConf()
	// 无论是否开启都前置条件都可触发
	for i, taskInfo := range m.getDbData().GetPreTaskInfos() {
		if i >= len(conf.GetPreCondition()) {
			break
		}
		preTaskConf := conf.GetPreCondition()[i]
		if f(preTaskConf, taskInfo) {
			m.commonSaveDB()
		}
	}

	// 只触发开启的活动
	if err := m.invalid(); err != nil {
		logInfo("活动不可用", zap.Error(err), zap.Int32("playerId", m.mgr.getPlayerId()), zap.Int64("activityId", m.getId()))
		return
	}
	m.rangeTemplates(func(template iTemplate) {
		template.rangeTasks(f)
	})
}

//
// getScoreReward
// @Description: 领取积分奖励
// @receiver m
// @param player
// @param index
// @return error
//
func (m *Activity) getScoreReward(player IPlayer, index int) error {
	conf := m.getConf()
	scoreSystem := conf.GetScoreSystem()
	if scoreSystem == nil {
		return errors.New("scoreSystem is nil")
	}
	if index >= len(scoreSystem) {
		return errors.New("scoreSystem index out")
	}
	scoreInfo := scoreSystem[index]

	if m.isGotScoreReward(index) {
		return errors.New("scoreSystem index got")
	}

	// 检测积分
	if err := player.OperateCheckCost([]*pb.ItemData{scoreInfo.GetScore()}); err != nil {
		return err
	}

	// 添加奖励
	if err := player.OperateAddReward(scoreInfo.GetReward()); err != nil {
		return err
	}

	m.setGotScoreRewardRecord(index)

	m.commonSaveDB()
	return nil
}

func (m *Activity) commonSaveDB() {
	m.callUpdateStatusFun(m.generateScoreUpdateDBData(), DataUpdate)
}

func (m *Activity) generateScoreUpdateDBData() *pb.OperateActivityDB {
	updateInfo := &pb.OperateActivityDB{
		ActivityId:   m.getId(),
		GotScores:    m.dbData.GetGotScores(),
		PreTaskInfos: m.dbData.GetPreTaskInfos(),
	}
	return updateInfo
}

//
// getSignTemplate
// @Description: 获取签到模板数据
// @receiver m
// @param index
// @return *signTemplate
//
func (m *Activity) getSignTemplate(index int) *signTemplate {
	template := m.getTemplate(index)
	if template == nil {
		return nil
	}
	if template.getType() != pb.ActivityTemplateType_SIGN_IN_TYPE {
		return nil
	}
	return template.(*signTemplate)
}

//
// getTaskTemplate
// @Description: 获取任务模板数据
// @receiver m
// @param index
// @return *taskTemplate
//
func (m *Activity) getTaskTemplate(index int) *taskTemplate {
	template := m.getTemplate(index)
	if template == nil {
		return nil
	}
	if template.getType() != pb.ActivityTemplateType_CONDITION_TYPE {
		return nil
	}
	return template.(*taskTemplate)
}

//
// getShopTemplate
// @Description: 获取商城模板数据
// @receiver m
// @param index
// @return *shopTemplate
//
func (m *Activity) getShopTemplate(index int) *shopTemplate {
	template := m.getTemplate(index)
	if template == nil {
		return nil
	}
	if template.getType() != pb.ActivityTemplateType_CONSUMPTION_TYPE {
		return nil
	}
	return template.(*shopTemplate)
}

//
// getCanReceiveReward
// @Description: 获取所有可领取奖励
// @receiver m
// @return []*ItemData
//
func (m *Activity) getCanReceiveReward(player IPlayer) []*pb.ItemData {
	var items []*pb.ItemData
	m.rangeTemplates(func(template iTemplate) {
		items = append(items, template.getCanReceiveReward()...)
	})
	items = append(items, m.getCanReceiveScoreReward(player)...)
	return items
}

//
// getCanReceiveScoreReward
// @Description: 获取可领取的积分奖励
// @receiver m
// @param player
// @return []*ItemData
//
func (m *Activity) getCanReceiveScoreReward(player IPlayer) []*pb.ItemData {
	conf := m.getConf()
	var items []*pb.ItemData
	for index, scoreInfo := range conf.GetScoreSystem() {
		if m.isGotScoreReward(index) {
			continue
		}
		// 检测积分
		if err := player.OperateCheckCost([]*pb.ItemData{scoreInfo.GetScore()}); err != nil {
			continue
		}
		items = append(items, scoreInfo.GetReward()...)
	}
	return items
}

//
// isGotScoreReward
// @Description: 是否已领取积分奖励
// @receiver m
// @param index
// @return bool
//
func (m *Activity) isGotScoreReward(index int) bool {
	dbData := m.getDbData()
	if dbData == nil {
		return true
	}
	gotScores := dbData.GetGotScores()
	_, ok := gotScores[int32(index)]
	return ok
}

//
// setGotScoreRewardRecord
// @Description: 保存领奖记录
// @receiver m
// @param index
//
func (m *Activity) setGotScoreRewardRecord(index int) {
	dbData := m.getDbData()
	if dbData == nil {
		return
	}
	gotScores := dbData.GetGotScores()
	if gotScores == nil {
		return
	}
	gotScores[int32(index)] = true
}

func (m *Activity) getClientData() *pb.Operate {
	return &pb.Operate{
		Detailed: m.getDbData(),
		Conf:     m.getConf(),
		Day:      m.openDay(),
	}
}
