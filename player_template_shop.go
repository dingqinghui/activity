/**
 * @Author: dingqinghui
 * @Description:商城模板
 * @File:  shop_template
 * @Version: 1.0.0
 * @Date: 2022/6/7 11:09
 */

package activity

import (
	"errors"
	"github.com/dingqinghui/activity/pb"
	"go.uber.org/zap"
)

func init() {
	registerTemplate(pb.ActivityTemplateType_CONSUMPTION_TYPE, newShopTemplate)
}

func newShopTemplate(day int32, index int32, conf *pb.ActivityTemplate, activity *Activity, dbData *pb.ActivityTemplateDB) iTemplate {
	if err := templateParameterCheck(conf, activity); err != nil {
		logError("newConditionTemplate", zap.Error(err))
		return nil
	}
	result := &shopTemplate{
		baseTemplate: newBaseTemplate(day, index, conf, activity, dbData),
	}

	result.init(result)
	return result
}

type shopTemplate struct {
	*baseTemplate
}

func (m *shopTemplate) initData() {
	m.dbData = &pb.ActivityTemplateDB{
		ConsumptionDB: &pb.ConsumptionTemplateDB{
			BuyCounts: make(map[int32]int32),
		},
	}
}

func (m *shopTemplate) getShopConf() *pb.ConsumptionTemplate {
	return m.conf.GetConsumption()
}

func (m *shopTemplate) getShopData() *pb.ConsumptionTemplateDB {
	return m.dbData.GetConsumptionDB()
}

func (m *shopTemplate) buy(player IPlayer, goodsIndex int) error {
	conf := m.getShopConf()
	if conf == nil {
		return errors.New("shop conf not exist")
	}

	dbData := m.getShopData()
	if dbData == nil {
		return errors.New("shop db data is nil")
	}

	if goodsIndex >= len(conf.GetSellGoods()) {
		return errors.New("shop conf goods not exist")
	}

	goodsConf := conf.GetSellGoods()[goodsIndex]

	buyCount := dbData.GetBuyCounts()[int32(goodsIndex)]

	// 限购
	if goodsConf.GetIsLimit() && buyCount >= goodsConf.GetLimitCount() {
		return errors.New("goods limit")
	}

	var costs []*pb.ItemData
	for _, item := range goodsConf.GetExpend() {
		costs = append(costs, &pb.ItemData{
			Id:  item.GetId(),
			Num: item.Num * goodsConf.GetDiscount() / 100,
		})
	}
	// 检测消耗
	if err := player.OperateCheckCost(costs); err != nil {
		return err
	}
	// 扣除消耗
	if err := player.OperateSubCost(costs); err != nil {
		return err
	}
	// 添加奖励
	if err := player.OperateAddReward(goodsConf.GetGoods()); err != nil {
		return err
	}

	if goodsConf.GetIsLimit() {
		dbData.GetBuyCounts()[int32(goodsIndex)] = buyCount + 1
	}
	m.saveDB()
	return nil
}

func (m *shopTemplate) saveDB() {
	m.activity.callUpdateStatusFun(m.generateUpdateData(), DataUpdate)
}

func (m *shopTemplate) generateUpdateData() *pb.OperateActivityDB {
	templateDB := &pb.ActivityTemplateDB{
		ConsumptionDB: m.getShopData(),
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
