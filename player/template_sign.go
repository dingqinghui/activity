/**
 * @Author: dingqinghui
 * @Description:签到模板
 * @File:  sign_template
 * @Version: 1.0.0
 * @Date: 2022/6/6 10:38
 */

package player

import (
	"errors"
	"github.com/dingqinghui/activity/global"
	"github.com/dingqinghui/activity/pb"
	"go.uber.org/zap"
	"math"
)

func init() {
	registerTemplate(pb.ActivityTemplateType_SIGN_IN_TYPE, newSignTemplate)
}

func newSignTemplate(day int32, index int32, conf *pb.ActivityTemplate, activity *Activity, dbData *pb.ActivityTemplateDB) iTemplate {
	if err := templateParameterCheck(conf, activity); err != nil {
		global.LogError("newConditionTemplate", zap.Error(err))
		return nil
	}
	result := &signTemplate{
		baseTemplate: newBaseTemplate(day, index, conf, activity, dbData),
	}
	result.init(result)
	return result
}

type signTemplate struct {
	*baseTemplate
}

//
// initData
// @Description: 初始化模板DB数据
// @receiver m
//
func (m *signTemplate) initData() {
	data := &pb.SignInTemplateDB{}
	conf := m.getSignConf()
	if conf == nil {
		return
	}
	for _, rules := range conf.GetRepairSignIn() {
		var taskList []*pb.OperateTaskInfo
		for range rules.GetRSI_Condition() {
			taskList = append(taskList, &pb.OperateTaskInfo{})
		}
		data.Conditions = append(data.Conditions, &pb.RepairCondition{Tasks: taskList})
	}
	m.dbData = &pb.ActivityTemplateDB{SignInDB: data}
}

//
// getSignConf
// @Description: 获取签到配置信息
// @receiver m
// @return *SignInTemplate
//
func (m *signTemplate) getSignConf() *pb.SignInTemplate {
	return m.conf.GetSignIn()
}

//
// getSignData
// @Description: 获取签到数据信息
// @receiver m
// @return *SignInTemplateDB
//
func (m *signTemplate) getSignData() *pb.SignInTemplateDB {
	return m.dbData.GetSignInDB()
}

//
// rangeTasks
// @Description: 遍历所有任务
// @receiver m
// @param f
//
func (m *signTemplate) rangeTasks(f RangeTaskFunType) {
	conf := m.getSignConf()
	if conf == nil {
		return
	}
	dbData := m.getSignData()
	if dbData == nil {
		return
	}
	needSaveDB := false
	rules := conf.GetRepairSignIn()
	for day, rule := range rules {
		for index, condConf := range rule.GetRSI_Condition() {
			if day >= len(dbData.GetConditions()) {
				continue
			}
			conditions := dbData.GetConditions()[day]
			if index >= len(conditions.GetTasks()) {
				continue
			}
			task := conditions.GetTasks()[index]
			if f(condConf, task) {
				needSaveDB = true
			}
		}
	}
	if needSaveDB {
		m.saveDB()
	}
}

//
// saveDB
// @Description: 调用回调进行模板数据存档
// @receiver m
//
func (m *signTemplate) saveDB() {
	m.activity.callUpdateStatusFun(m.generateUpdateData(), DataUpdate)
}

//
// generateUpdateData
// @Description: 生成存档数据
// @receiver m
// @return *OperateActivityDB
//
func (m *signTemplate) generateUpdateData() *pb.OperateActivityDB {
	templateDB := &pb.ActivityTemplateDB{
		SignInDB: m.getSignData(),
	}
	list := &pb.ActivityDBList{
		List: map[int32]*pb.ActivityTemplateDB{m.getIndex(): templateDB},
	}
	updateInfo := &pb.OperateActivityDB{
		ActivityId:   m.activity.getId(),
		ActivityList: map[int32]*pb.ActivityDBList{m.getDay(): list},
	}
	return updateInfo
}

func (m *signTemplate) isLoginTrigger() bool {
	conf := m.getSignConf()
	if conf == nil {
		return false
	}
	return !conf.GetTriggerCondition()
}

func (m *signTemplate) canSignDay() int32 {
	day := m.activity.openDay()
	return day + 1
}

func (m *signTemplate) sign(player IPlayer) error {
	// 检测签到条件
	if err := m.checkSignCondition(player); err != nil {
		return err
	}

	// 下发奖励
	if err := m.addSignReward(player); err != nil {
		return err
	}

	// 签到
	dbData := m.getSignData()
	if dbData == nil {
		return errors.New("sign db data is nil")
	}
	dbData.SignedDay += 1
	dbData.LastSignTimestamp = global.NowTimestamp()
	m.saveDB()
	global.LogInfo("签到成功", zap.Int32("playerId", player.GetId()), zap.Int64("activityId", m.activity.getId()), zap.Int32("signedDay", dbData.GetSignedDay()))
	return nil
}

func (m *signTemplate) checkSignCondition(player IPlayer) error {
	conf := m.getSignConf()
	if conf == nil {
		return errors.New("sign conf is nil")
	}
	dbData := m.getSignData()
	if dbData == nil {
		return errors.New("sign db data is nil")
	}

	if !global.IsDifferDay(global.NowTimestamp(), dbData.GetLastSignTimestamp()) {
		global.LogError("今日已签到", zap.Int32("playerId", player.GetId()))
		return errors.New("today signed")
	}

	if dbData.GetSignedDay() >= m.getCanSignCount() {
		return errors.New("sign count limit")
	}
	return nil
}

func (m *signTemplate) repair(player IPlayer) error {
	if err := m.repairCondition(player); err != nil {
		global.LogError("补签失败，条件检测失败", zap.Int32("playerId", player.GetId()))
		return errors.New("repair sign condition ")
	}

	if err := m.addSignReward(player); err != nil {
		global.LogError("补签失败，添加奖励失败", zap.Int32("playerId", player.GetId()), zap.Error(err))
		return err
	}

	// 签到
	dbData := m.getSignData()
	if dbData == nil {
		return errors.New("sign db data is nil")
	}
	dbData.SignedDay += 1
	dbData.RepairCount += 1
	m.saveDB()

	global.LogInfo("补签成功", zap.Int32("playerId", player.GetId()), zap.Int64("activityId", m.activity.getId()),
		zap.Int32("signedDay", dbData.GetSignedDay()), zap.Int32("repairCount", dbData.GetRepairCount()))
	return nil
}

func (m *signTemplate) getCanSignCount() int32 {
	conf := m.getSignConf()
	if conf == nil {
		return 0
	}
	dbData := m.getSignData()
	if dbData == nil {
		return 0
	}
	return int32(math.Min(float64(m.canSignDay()), float64(conf.GetSignInCount())))
}

//
// getRepairMaxCount
// @Description: 获取可以补签的最大次数
// @receiver m
// @return int32
//
func (m *signTemplate) getCantRepairCount() int32 {
	conf := m.getSignConf()
	if conf == nil {
		return 0
	}
	dbData := m.getSignData()
	if dbData == nil {
		return 0
	}
	return int32(math.Min(float64(m.canSignDay()-1), float64(conf.GetRepairSignInCount())))
}

//
// addSignReward
// @Description: 添加签到奖励
// @receiver m
//
func (m *signTemplate) addSignReward(player IPlayer) error {
	conf := m.getSignConf()
	if conf == nil {
		return errors.New("sign conf is nil")
	}
	dbData := m.getSignData()
	if dbData == nil {
		return errors.New("sign db data is nil")
	}

	// 下发奖励
	rewards := conf.GetRewardList()
	if int(dbData.GetSignedDay()) < len(rewards) {
		reward := rewards[dbData.GetSignedDay()]
		if err := player.OperateAddReward(reward.GetSignInReward()); err != nil {
			global.LogError("签到失败",
				zap.Int32("playerId", player.GetId()),
				zap.Int64("activityId", m.activity.getId()),
				zap.Int32("signedDay", dbData.GetSignedDay()),
				zap.Error(err),
			)
			return errors.New("repair sign add reward fail ")
		}
	}
	return nil
}

func (m *signTemplate) repairCondition(player IPlayer) error {
	conf := m.getSignConf()
	if conf == nil {
		return errors.New("sign conf is nil")
	}
	dbData := m.getSignData()
	if dbData == nil {
		return errors.New("sign db data is nil")
	}

	// 已补签次数 > 可补签次数
	if dbData.GetRepairCount() >= m.getCantRepairCount() {
		return errors.New("repair sign count limit")
	}

	// 已签到天数
	signedDay := dbData.GetSignedDay()
	// 补签条件
	rules := conf.GetRepairSignIn()
	// 无条件
	if int(signedDay) >= len(rules) {
		return nil
	}
	// 检测补签条件
	rule := rules[signedDay]
	// 道具消耗补签
	if rule.GetRSI_Expend() != nil {
		if err := player.OperateSubCost(rule.GetRSI_Expend()); err != nil {
			global.LogError("补签失败，道具不足", zap.Int32("playerId", player.GetId()), zap.Int32("signedDay", signedDay))
			return errors.New("repair condition not enough expend")
		}
		return nil
	}

	// 无条件
	if int(signedDay) >= len(dbData.GetConditions()) {
		return nil
	}

	// 检测补签条件
	condition := dbData.GetConditions()[signedDay]
	if condition == nil {
		return nil
	}
	// 任务条件
	if rule.GetRSI_Condition() != nil {
		for _, task := range condition.GetTasks() {
			if task.GetTaskState() == pb.OperateTaskState_OTS_Doing {
				global.LogError("补签失败，条件不满足", zap.Int32("playerId", player.GetId()), zap.Int32("signedDay", signedDay))
				return errors.New("repair condition task not finish")
			}
		}
	}
	return nil
}
