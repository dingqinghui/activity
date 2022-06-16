/**
 * @Author: dingqinghui
 * @Description:条件模板
 * @File:  condition_template
 * @Version: 1.0.0
 * @Date: 2022/6/7 10:30
 */

package player

import (
	"errors"
	"github.com/dingqinghui/activity/global"
	"github.com/dingqinghui/activity/pb"
	"go.uber.org/zap"
)

func init() {
	registerTemplate(pb.ActivityTemplateType_CONDITION_TYPE, newConditionTemplate)
}

func newConditionTemplate(day int32, index int32, conf *pb.ActivityTemplate, activity *Activity, dbData *pb.ActivityTemplateDB) iTemplate {
	if err := templateParameterCheck(conf, activity); err != nil {
		global.LogError("newConditionTemplate", zap.Error(err))
		return nil
	}
	result := &taskTemplate{
		baseTemplate: newBaseTemplate(day, index, conf, activity, dbData),
	}
	result.init(result)
	return result
}

type taskTemplate struct {
	*baseTemplate
}

func (m *taskTemplate) initData() {
	var taskList []*pb.OperateTaskInfo
	conf := m.getTaskConf()
	if conf == nil {
		return
	}
	for range conf.GetData() {
		taskList = append(taskList, &pb.OperateTaskInfo{})
	}

	m.dbData = &pb.ActivityTemplateDB{
		ConditionDB: &pb.ConditionTemplateDB{
			TaskInfo: taskList,
		},
	}
}

func (m *taskTemplate) getTaskConf() *pb.ConditionTemplate {
	return m.conf.GetCondition()
}

func (m *taskTemplate) getTaskData() *pb.ConditionTemplateDB {
	return m.dbData.GetConditionDB()
}

func (m *taskTemplate) getTaskInfo(taskId int32) *pb.OperateTaskInfo {
	condition := m.dbData.GetConditionDB()
	if condition == nil {
		return nil
	}
	if int(taskId) >= len(condition.GetTaskInfo()) {
		return nil
	}
	return condition.GetTaskInfo()[taskId]
}

func (m *taskTemplate) finishTask(player IPlayer, taskId int32) error {
	template := m.getTaskConf()
	if template == nil {
		return errors.New("template is nil")
	}
	taskList := template.GetData()
	if taskList == nil {
		return errors.New("taskList is nil")
	}

	if len(taskList) <= int(taskId) {
		return errors.New("taskId out range")
	}
	condition := taskList[taskId]
	if condition == nil {
		return errors.New("task not exist")
	}

	task := m.getTaskInfo(taskId)
	if task == nil {
		return errors.New("task is nil")
	}

	if task.GetTaskState() != pb.OperateTaskState_OTS_Finish {
		return errors.New("task status err")
	}

	if err := player.OperateAddReward(condition.GetRewardList()); err != nil {
		return err
	}

	task.TaskState = pb.OperateTaskState_OTS_Over

	m.saveDB()
	return nil
}

func (m *taskTemplate) rangeTasks(f RangeTaskFunType) {
	if f == nil {
		return
	}
	conf := m.getTaskConf().GetData()
	taskInfos := m.getTaskData()
	needSaveDB := false
	for index, taskConf := range conf {
		if index >= len(taskInfos.GetTaskInfo()) {
			continue
		}
		task := taskInfos.GetTaskInfo()[index]
		if f(taskConf, task) {
			needSaveDB = true
		}
	}
	if needSaveDB {
		m.saveDB()
	}
}

func (m *taskTemplate) getCanReceiveReward() []*pb.ItemData {
	var items []*pb.ItemData

	condition := m.dbData.GetConditionDB()
	if condition == nil {
		return nil
	}
	conf := m.getTaskConf()
	if conf == nil {
		return nil
	}

	for index, task := range condition.GetTaskInfo() {
		if index >= len(conf.GetData()) {
			break
		}
		taskConf := conf.GetData()[index]
		if task.GetTaskState() == pb.OperateTaskState_OTS_Finish {
			items = append(items, taskConf.GetRewardList()...)
		}
	}
	return items
}

func (m *taskTemplate) saveDB() {
	m.activity.callUpdateStatusFun(m.generateUpdateData(), global.DataUpdate)
}

func (m *taskTemplate) generateUpdateData() *pb.OperateActivityDB {
	templateDB := &pb.ActivityTemplateDB{
		ConditionDB: m.getTaskData(),
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
