package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime/pprof"
	"time"

	"github.com/adamluo159/cellnet"
	"github.com/adamluo159/cellnet/codec"
	_ "github.com/adamluo159/cellnet/codec/json"
	"github.com/adamluo159/cellnet/peer"
	_ "github.com/adamluo159/cellnet/peer/tcp"
	"github.com/adamluo159/cellnet/proc"
	_ "github.com/adamluo159/cellnet/proc/tcp"
	"github.com/adamluo159/cellnet/util"
	"github.com/davyxu/golog"
)

func server() {
	queue := cellnet.NewEventQueue()

	p := peer.NewGenericPeer("tcp.Acceptor", "server", "127.0.0.1:7701", queue)

	dispatcher := proc.NewMessageDispatcherBindPeer(p, "tcp.ltv")

	dispatcher.RegisterMessage("main.TestEchoACK", func(ev cellnet.Event) {

		msg := ev.Message().(*TestEchoACK)

		ev.Session().Send(&TestEchoACK{
			Msg:   msg.Msg,
			Value: msg.Value,
		})
	})

	p.Start()

	queue.StartLoop()
}

func client() {

	queue := cellnet.NewEventQueue()

	p := peer.NewGenericPeer("tcp.Connector", "client", "127.0.0.1:7701", queue)

	rv := proc.NewSyncReceiver(p)

	proc.BindProcessorHandler(p, "tcp.ltv", rv.EventCallback())

	p.Start()

	queue.StartLoop()

	rv.WaitMessage("cellnet.SessionConnected")

	p.(cellnet.TCPConnector).Session().Send(&TestEchoACK{
		Msg:   "hello",
		Value: 1234,
	})

	begin := time.Now()

	var lastcheck time.Time

	const total = 10 * time.Second

	for {

		now := time.Now()

		if now.Sub(begin) >= total {
			break
		}

		if now.Sub(lastcheck) >= time.Second {
			fmt.Printf("progress: %d%%\n", now.Sub(begin)*100/total)
			lastcheck = now
		}

		rv.Recv(func(ev cellnet.Event) {

			ev.Session().Send(&TestEchoACK{
				Msg:   "hello",
				Value: 1234,
			})

		})
	}

}

var profile = flag.String("profile", "", "write cpu profile to file")

type TestEchoACK struct {
	Msg   string
	Value int32
}

func (self *TestEchoACK) String() string { return fmt.Sprintf("%+v", *self) }

func init() {
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("json"),
		Type:  reflect.TypeOf((*TestEchoACK)(nil)).Elem(),
		ID:    int(util.StringHash("main.TestEchoACK")),
	})
}

// go build -o bench.exe main.go
// ./bench.exe -profile=mem.pprof
// go tool pprof -alloc_space -top bench.exe mem.pprof
func main() {

	flag.Parse()

	f, err := os.Create(*profile)
	if err != nil {
		log.Println(*profile)
	}

	golog.SetLevelByString("tcpproc", "info")

	server()

	client()

	if f != nil {
		pprof.WriteHeapProfile(f)
		f.Close()
	}

}
