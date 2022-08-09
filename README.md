

### 模块组成

ActivityMgr：管理玩家所有活动数据

Activity：单个活动实例，管理多个模板

templateMgr：模板工厂，提供模板注册和创建

iTemplate：模板接口定义

baseTemplate：模板基类实现

shopTemplate：商城模板实现

taskTemplate：任务模板实现

signTemplate：签到模板实现

operatorActivityMgr：

全局活动管理器，负责管理所有Gm运营活动数据
当GM通知logic添加/删除活动数据时，由logic调用Add/Delete接口，将数据添加到管理器中，管理器添加成功后，会通过回调函数通知logic，logic可进行二次处理。如果活动到期也会通过回调方式通知logic。


logger：

日志处理器，使用Uber-go Zap实现
支持两种构造方式
由外部直接传入zap.Logger对象（WithLogger）
外部传入logPath 和 logLevel，库构创建zap.Logger对象（WithLogConfig）



pb :

pb文件夹下包含活动pb数据结构定义，以及与客户端通信协议





### 接入

#### 全局数据

全局数据可以理解为活动配置，通过GM进行配置，存储于Redis，GM使用redis 发布订阅通知服务器数据变更。服务器收到数据变更后调用activity.Add,activity.Delete,改变运营活动库中数据。运营活动库数据发生变化通过回调方式通知上层，上层对改变的活动进行DB处理。

```go
//
// GlobalActivityDataUpdate
// @Description: 全局数据更改回调函数
// @param activity
// @param cmd
//
func GlobalActivityDataUpdate(activity *pb.OperateActivity, cmd DataCmd) {
	switch cmd {
	case DataAdd:
		
	case DataDelete:
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
	return nowTimestamp()
}

// 全局数据初始化
Init(nil, GlobalActivityDataUpdate, GetAreaStartTime, nil)
Init(nil, GlobalActivityDataUpdate, GetAreaStartTime, WithLogger(zap.New(zapcore.NewTee())))
Init(nil, GlobalActivityDataUpdate, GetAreaStartTime, WithLogConfig("", zap.DebugLevel))

// 添加活动
Add(&pb.OperateActivity{})
// 删除活动
Delete(1)
// 设置时区 默认东八区
SetTimeZero(8)
// 设置每日更新时间(每日几点算跨天)
SetEverydayUpdateHour(8)
```

![image-20220808102515613](https://s2.loli.net/2022/08/08/mESJhytX3DQ8Pir.png)


#### 玩家数据

##### 定义player

player实现 IPlayer 接口，实现玩家的奖励发放，消耗扣除等逻辑。

```go
func newPlayer() *player {
   p := &player{}
   p.tick()
   return p
}

type player struct {
   operate *PlayerActivityMgr
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
```

同时要周期性的调用PlayerActivityMgr.CheckNewAndDelete,此接口会检测是否有新添加的活动，以及活动是否过期。检测到添加/删除活动则通过回调函数进行通知。



##### 创建玩家活动数据模块

```go
func (p *player) GetOperate() *PlayerActivityMgr {
   if p.operate == nil {
      // 创建玩家运营活动模块  (区服，渠道，注册时间，数据变更回调，初始化活动数据)
      p.operate = NewPlayerActivityMgr(p, 101, 10001, nowTimestamp(),
         PlayerActivityDataUpdate, nil)
   }
   return p.operate
}
```
##### 活动数据变更回调

回调时需要注意当cmd == DataAdd时，updateInfo为活动完整DB数据，当cmd == DataUpdate，updateInfo为活动更改数据,未更改的数据赋值为nil

```go
func PlayerActivityDataUpdate(playerId int32, activityId int64, cmd DataCmd, updateInfo *pb.OperateActivityDB) {
   switch cmd {
   case DataAdd, DataUpdate:
      // 更新玩家db数据
   case DataDelete:
      // 删除玩家db数据
   default:
   }
}
```

![image-20220808102502069](https://s2.loli.net/2022/08/08/dorcLpSP6fzGCxb.png)

详细测试代码见 activity_test.go





