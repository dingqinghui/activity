/**
 * @Author: dingQingHui
 * @Description:活动模板基类
 * @File: player_base_template
 * @Version: 1.0.0
 * @Date: 2022/6/15 14:52
 */

package player

import (
	"errors"
	"github.com/dingqinghui/activity/pb"
)

func newBaseTemplate(day int32, index int32, conf *pb.ActivityTemplate, activity *Activity, dbData *pb.ActivityTemplateDB) *baseTemplate {
	return &baseTemplate{
		day:      day,
		index:    index,
		conf:     conf,
		activity: activity,
		dbData:   dbData,
	}
}

//
// iTemplate
// @Description: 活动模板
//
type iTemplate interface {
	getDay() int32
	getIndex() int32
	getType() pb.ActivityTemplateType
	getCanReceiveReward() []*pb.ItemData
	rangeTasks(f RangeTaskFunType)
	getDbData() *pb.ActivityTemplateDB
	initData()
}

type baseTemplate struct {
	//
	// day
	// @Description: 模板所属天
	//
	day int32
	//
	// index
	// @Description: 模板所属索引
	//
	index int32

	//
	//  config
	// @Description: 活动模板Gm配置信息
	//
	conf *pb.ActivityTemplate
	//
	// activity
	// @Description: 活动数据
	//
	activity *Activity

	//
	// dbData
	// @Description: 存档数据
	//
	dbData *pb.ActivityTemplateDB
}

func (m *baseTemplate) init(template iTemplate) {
	if m.dbData == nil {
		template.initData()
	}
}

func (m *baseTemplate) initData() {

}

//
// Type
// @Description: 模板类型
// @receiver m
// @return ActivityTemplateType
//
func (m *baseTemplate) getType() pb.ActivityTemplateType {
	return m.conf.TemplateType
}

func (m *baseTemplate) getCanReceiveReward() []*pb.ItemData {
	return nil
}

func (m *baseTemplate) getDay() int32 {
	return m.day
}
func (m *baseTemplate) getIndex() int32 {
	return m.index
}
func (m *baseTemplate) rangeTasks(_ RangeTaskFunType) {
}

func (m *baseTemplate) getDbData() *pb.ActivityTemplateDB {
	return m.dbData
}

func templateParameterCheck(data *pb.ActivityTemplate, activity *Activity) error {
	if data == nil {
		return errors.New("data is nil ")
	}
	if activity == nil {
		return errors.New("activity data is nil ")
	}
	return nil
}
