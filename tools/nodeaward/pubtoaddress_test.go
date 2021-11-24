package main

import (
	"encoding/hex"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/system/p2p/dht"
	"testing"
)

func Test_PubtoAddress(t *testing.T){
	pubstr, err := dht.PeerIDToPubkey("16Uiu2HAmKkaST91Nx9pWVDxFZxoQQze89ZESUEQhcjsKPJXQag1u")
	if err!=nil{
		return
	}
	pub,_:=hex.DecodeString(pubstr)
	t.Log("address:",address.PubKeyToAddress(pub).String())
}