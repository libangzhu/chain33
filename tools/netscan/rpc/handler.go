package rpc

import (
	"errors"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	ctypes "github.com/33cn/chain33/types"
	"time"
)
var  NetRatePrefix="level-netrate-"

type NetScan struct {
	client queue.Client
}

type ReqNil struct {
}

//NetRateReq 一次性返回当天所有的统计数据
type NetRateReq struct {
	Date string `json:"date,omitempty"`
}
const(
	EventPeerLocaltionInfo=112
	EventPeersNetRateInfo=113
)
func (n *NetScan) PeersLocation(in *ReqNil, result *interface{}) error {
	log.Info("PeersLocation")
	msg := n.client.NewMessage("p2p", EventPeerLocaltionInfo, nil)
	err := n.client.SendTimeout(msg, true, time.Second*15)
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

func (n *NetScan)PeersNetRate(in *NetRateReq,result *interface{})error{

	msg := n.client.NewMessage("p2p", EventPeersNetRateInfo, in)
	err := n.client.SendTimeout(msg, true, time.Second*10)
	if err != nil {
		log.Error("PeersLocation", "Error", err.Error())
		return err
	}
	resp, err := n.client.WaitTimeout(msg, time.Second*2)
	if err != nil {
		return err
	}
	var ok bool
	*result,ok = resp.GetData().(*PeersNetRateResp)
	if !ok{
		resp := resp.GetData().(*ctypes.Reply)
		return errors.New(string(resp.GetMsg()))

	}

	return nil
}


type ConriesInfo struct{
	Countries []*CountryInfo `json:"countries"`
}

type CountryInfo struct{
	Country string  `json:"country"`
	Value int `json:"value"`
}

type PeersNetRateResp struct {
	Netrate []*NetrateInfo `json:"netrate"`
}

type NetrateInfo struct {
	Time string `json:"time"`
	Rate string `json:"rate"`
}