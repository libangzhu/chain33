package p2pnext

import (
	"sync"
	"time"
)

var monitor *Monitor

/**
MsgId 消息ID
Timestamp 消息生成时间
TTL 超时时间
 */
type MessageInfo struct {
	Topic string
	MsgId string
	Timestamp  int64
	TTL int64
}

type Monitor struct{
	node     *Node     // local host
	registry sync.Map  // 保存登记的消息
}

// NewMonitor produce a monitor object
func NewMonitor(n *Node) *Monitor {
	if monitor == nil {
		monitor = &Monitor{node:n}
	}

	return  monitor
}

// 注册需要进行监控的消息的topic 返回超时消息的chan
func RegisterMonitor(topic string) chan interface{} {
	timeoutMsgChan := monitor.node.Pubsub.Sub(topic)
	return timeoutMsgChan
}

func RegisterMessage(topic, msgId string, timestamp, timeout int64 ) bool {
	msgInfo := &MessageInfo{
		Topic: topic,
		MsgId: msgId,
		Timestamp: timestamp,
		TTL: timeout,
	}
	monitor.registry.Store(msgId, msgInfo)
	return true
}

func (m *Monitor) StartMonitor()  {
	ticker := time.NewTicker(1 *time.Second)
	defer  ticker.Stop()
	for {
		select {
		case <- ticker.C:
			m.registry.Range(func(key, value interface{}) bool {
				msginfo := value.(*MessageInfo)
				if msginfo.Timestamp + msginfo.TTL < time.Now().Unix() {
					m.node.Pubsub.FIFOPub(msginfo, msginfo.Topic)
					m.registry.Delete(msginfo.MsgId)
				}
				return true
			})
		}
	}
}
