/**
 * @Author: dingqinghui
 * @Description:活动模板管理器
 * @File:  template_mgr
 * @Version: 1.0.0
 * @Date: 2022/6/1 16:17
 */

package activity

import (
	"github.com/dingqinghui/activity/pb"
	"sync"
)

type newTemplateFunc func(day int32, index int32, conf *pb.ActivityTemplate, activity *Activity, dbData *pb.ActivityTemplateDB) iTemplate

var (
	tm              *templateMgr
	onceTemplateMgr sync.Once
)

func getTemplateMgr() *templateMgr {
	onceTemplateMgr.Do(func() {
		tm = &templateMgr{}
	})
	return tm
}

type templateMgr struct {
	sync.Map
}

func (m *templateMgr) register(tt pb.ActivityTemplateType, f newTemplateFunc) {
	m.Store(tt, f)
}
func (m *templateMgr) newTemplate(day int32, index int32, conf *pb.ActivityTemplate, activity *Activity, dbData *pb.ActivityTemplateDB) iTemplate {
	v, ok := m.Load(conf.GetTemplateType())
	if !ok {
		return nil
	}
	f, ok := v.(newTemplateFunc)
	if f == nil || !ok {
		return nil
	}
	return f(day, index, conf, activity, dbData)
}

func newTemplate(day int32, index int32, conf *pb.ActivityTemplate, activity *Activity, dbData *pb.ActivityTemplateDB) iTemplate {
	if conf == nil {
		return nil
	}
	return getTemplateMgr().newTemplate(day, index, conf, activity, dbData)
}

func registerTemplate(tt pb.ActivityTemplateType, f newTemplateFunc) {
	getTemplateMgr().register(tt, f)
}
