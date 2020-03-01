package basecell

import (
	"fmt"
	"reflect"
	"time"

	"github.com/davyxu/golog"

	"github.com/adamluo159/cellnet"
)

var log *golog.Logger = nil //= golog.New("websocket_bceller")
//DefaultCell 默认服务
var DefaultCell *BaseCell = nil

//IModule 模块接口
type iModule interface {
	Init()
	Name() string
	OnDestory()
}

//BaseCell 基础服务
type BaseCell struct {
	CodecName string //编码名

	bcellName    string //服务名字
	modules      []iModule
	msgHandler   map[reflect.Type]func(ev cellnet.Event)
	eventHandler map[reflect.Type]func(ev interface{}) interface{}
	queue        cellnet.EventQueue
	peer         cellnet.GenericPeer
}

//SetLog 设置日志
func SetLog(l *golog.Logger) {
	log = l
}

//Start 服务开始
func (bcell *BaseCell) Start(mods ...iModule) {
	tmpNames := []string{}
	for _, m := range mods {
		for _, name := range tmpNames {
			if name == m.Name() {
				panic(fmt.Sprintf("repeat module name:%s", m.Name()))
			}
		}
		m.Init()
		tmpNames = append(tmpNames, m.Name())
	}
	bcell.modules = mods

	bcell.queue.EnableCapturePanic(true)

	// 开始侦听
	bcell.peer.Start()

	// 事件队列开始循环
	bcell.queue.StartLoop()
}

//Stop 服务停止
func (bcell *BaseCell) Stop() {
	bcell.peer.Stop()
	bcell.queue.StopLoop()

	bcell.queue.Wait()

	for _, m := range bcell.modules {
		m.OnDestory()
	}
}

//RegitserMessage 注册默认消息响应
func RegitserMessage(msg interface{}, f func(ev cellnet.Event)) {
	if DefaultCell == nil {
		panic("RegitserModuleMsg Default nil")
	}
	DefaultCell.RegisterMessage(msg, f)
}

//RegitserEvent 注册默认消息响应
func RegitserEvent(msg interface{}, f func(ev interface{}) interface{}) {
	if DefaultCell == nil {
		panic("RegitserModuleEvt Default nil")
	}
	DefaultCell.RegisterEvent(msg, f)
}

//RegisterMessage 注册消息回调
func (bcell *BaseCell) RegisterMessage(msg interface{}, f func(ev cellnet.Event)) {
	bcell.msgHandler[reflect.TypeOf(msg)] = f
}

//RegisterEvent 注册事件消息回调
func (bcell *BaseCell) RegisterEvent(evt interface{}, f func(ev interface{}) interface{}) {
	bcell.eventHandler[reflect.TypeOf(evt)] = f
}

//PostEvent 事件推送
func (bcell *BaseCell) PostEvent(evt interface{}) {
	bcell.queue.Post(func() {
		f, ok := bcell.eventHandler[reflect.TypeOf(evt)]
		if ok {
			f(evt)
		} else {
			log.Errorln("PostEvent not found event handler ", reflect.TypeOf(evt))
		}
	})
}

//PostEventSync  同步事件推送
func (bcell *BaseCell) PostEventSync(evt interface{}) interface{} {
	ch := make(chan interface{}, 1)

	bcell.queue.Post(func() {
		f, ok := bcell.eventHandler[reflect.TypeOf(evt)]
		if ok {
			ch <- f(evt)
		} else {
			log.Errorln("PostEventSync not found event handler ", reflect.TypeOf(evt))
		}
	})

	// 等待RPC回复
	select {
	case v := <-ch:
		return v
	case <-time.After(time.Second * 15):
		return nil
	}
}

//PostEventAsync 异步调用
func (bcell *BaseCell) PostEventAsync(evt interface{}, cb func(ret interface{})) {
	bcell.queue.Post(func() {
		f, ok := bcell.eventHandler[reflect.TypeOf(evt)]
		if ok {
			cb(f(evt))
		} else {
			log.Errorln("PostEventAsync not found event handler ", reflect.TypeOf(evt))
		}
	})
}
