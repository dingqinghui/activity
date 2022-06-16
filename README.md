运营活动管理



使用见 activity_test.go




operatorActivityMgr：

全局活动管理器，负责管理所有Gm运营活动数据
当GM通知logic添加/删除活动数据时，由logic调用Add/Delete接口，将数据添加到管理器中，管理器添加成功后，会通过回调函数通知logic，logic可进行二次处理。如果活动到期也会通过回调方式通知logic。




logger：

日志处理器，使用Uber-go Zap实现
支持两种构造方式
由外部直接传入zap.Logger对象（WithLogger）
外部传入logPath 和 logLevel，库构创建zap.Logger对象（WithLogConfig）



ActivityMgr：管理玩家所有活动数据

Activity：单个活动实例，管理多个模板

templateMgr：模板工厂，提供模板注册和创建

iTemplate：模板接口定义

baseTemplate：模板基类实现

shopTemplate：商城模板实现

taskTemplate：任务模板实现

signTemplate：签到模板实现
