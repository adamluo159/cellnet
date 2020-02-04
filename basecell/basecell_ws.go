package basecell

import (
	"reflect"
	"time"

	"github.com/adamluo159/cellnet"
	"github.com/adamluo159/cellnet/peer"
	"github.com/adamluo159/cellnet/proc"
)

//NewWsServer 创建ws服务器
func NewWsServer(addr string, name string) *BaseCell {
	queue := cellnet.NewEventQueue()

	bcell := &BaseCell{
		bcellName:    name,
		queue:        queue,
		peer:         peer.NewGenericPeer("gorillaws.Acceptor", name, addr, queue),
		msgHandler:   make(map[reflect.Type]func(ev cellnet.Event)),
		eventHandler: make(map[reflect.Type]func(ev interface{}) interface{}),
		CodecName:    "gogopb",
	}
	proc.BindProcessorHandler(bcell.peer, "gorillaws.ltv", func(ev cellnet.Event) {
		f, ok := bcell.msgHandler[reflect.TypeOf(ev.Message())]
		if ok {
			f(ev)
		} else {
			log.Errorln("ws Server not found message handler ", ev.Message())
		}
	})

	if DefaultCell == nil {
		DefaultCell = bcell
	}

	return bcell
}

//NewWsClient 创建客户端
func NewWsClient(addr string, name string) *BaseCell {
	queue := cellnet.NewEventQueue()

	p := peer.NewGenericPeer("gorillaws.Connector", "client", addr, queue)
	p.(cellnet.WSConnector).SetReconnectDuration(time.Second * 5)

	bcell := &BaseCell{
		bcellName:  name,
		queue:      queue,
		peer:       p,
		msgHandler: make(map[reflect.Type]func(ev cellnet.Event)),
		CodecName:  "gogopb",
	}

	proc.BindProcessorHandler(p, "gorillaws.ltv", func(ev cellnet.Event) {
		f, ok := bcell.msgHandler[reflect.TypeOf(ev.Message())]
		if ok {
			f(ev)
		} else {
			log.Errorln("client not found message handler ", reflect.TypeOf(ev.Message()))
		}
	})

	return bcell
}
