package rpc

import (
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	"time"
)

type NetScan struct {
	client queue.Client
}

type ReqNil struct {
}
const(
	EventPeerLocaltionInfo=112
)
func (n *NetScan) PeersLocation(in *ReqNil, result *interface{}) error {
	log.Info("PeersLocation")
	msg := n.client.NewMessage("p2p", EventPeerLocaltionInfo, nil)
	err := n.client.SendTimeout(msg, true, time.Second*2)
	if err != nil {
		log.Error("PeersLocation", "Error", err.Error())
		return err
	}
	resp, err := n.client.WaitTimeout(msg, time.Second*2)
	if err != nil {
		return err
	}

	*result = resp.GetData().(*ConriesInfo)
	return nil
}


type ConriesInfo struct{
	Countries []*CountryInfo `json:"countries"`
}
type CountryInfo struct{
	Country string  `json:"country"`
	Value int `json:"value"`
}