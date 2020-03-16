package basecell

import (
	"time"

	"github.com/adamluo159/cellnet"
	"github.com/adamluo159/cellnet/peer"
	"github.com/adamluo159/cellnet/proc"
)

//NewWsServer 创建ws服务器
func (bcell *BaseCell) NewWsServer(addr string) {
	bcell.peer = peer.NewGenericPeer("gorillaws.Acceptor", "", addr, bcell.queue)
	proc.BindProcessorHandler(bcell.peer, "gorillaws.ltv", bcell.msgQueue())
}

//NewWsClient 创建ws客户端
func (bcell *BaseCell) NewWsClient(addr string) {
	bcell.peer = peer.NewGenericPeer("gorillaws.Connector", "client", addr, bcell.queue)
	bcell.peer.(cellnet.WSConnector).SetReconnectDuration(time.Second * 5)
	proc.BindProcessorHandler(bcell.peer, "gorillaws.ltv", bcell.msgQueue())
}
