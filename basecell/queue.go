package basecell

import (
	"hash/crc32"
	"time"

	"github.com/adamluo159/cellnet"
)

//GetQidByKey 获取qid
func GetQidByKey(key string) int {
	return DefaultCell.GetQidByKey(key)
}

//GetQidByKey 获取qid
func (bcell *BaseCell) GetQidByKey(key string) int {
	v := int(crc32.ChecksumIEEE([]byte(key)))
	if v < 0 {
		return (-v) % bcell.MsgQueueLen
	}
	return v % bcell.MsgQueueLen
}

func (bcell *BaseCell) getQueue(qid int) cellnet.EventQueue {
	q := bcell.queue
	if bcell.MsgQueueLen > 0 {
		if qid > bcell.MsgQueueLen-1 {
			log.Errorln("qid < 0 ")
			return nil
		}
		if qid < 0 {
			log.Errorf("qid < 0")
			return nil
		}
		q = bcell.queues[qid]
	}
	return q
}

//Post 事件推送
func Post(qid int, f func()) {
	DefaultCell.Post(qid, f)
}

//Post 事件推送
func (bcell *BaseCell) Post(qid int, f func()) {
	bcell.getQueue(qid).Post(f)
}

//PostSync 投递
func PostSync(qid int, f func() interface{}) interface{} {
	return DefaultCell.PostSync(qid, f)
}

//PostSync 投递
func (bcell *BaseCell) PostSync(qid int, f func() interface{}) interface{} {
	ch := make(chan interface{}, 1)
	bcell.getQueue(qid).Post(func() {
		ch <- f()
	})
	// 等待RPC回复
	select {
	case v := <-ch:
		return v
	case <-time.After(time.Second * 30):
		return nil
	}
}

//PostAsync 异步调用
func PostAsync(qid int, f func() interface{}, cb func(ret interface{})) {
	DefaultCell.PostAsync(qid, f, cb)
}

//PostAsync 异步调用
func (bcell *BaseCell) PostAsync(qid int, f func() interface{}, cb func(ret interface{})) {
	bcell.getQueue(qid).Post(func() { cb(f()) })
}
