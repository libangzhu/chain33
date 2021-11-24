package main

import (
	"flag"
	logger "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	p2pty "github.com/33cn/chain33/system/p2p/dht/types"
	"github.com/33cn/chain33/tools/netscan/rpc"
)

var log = logger.New("module", p2pty.DHTTypeName)
var (
	path       = flag.String("p", "./datadir", "scan data path")
	gossipPath = flag.String("gop", "", "gossip data path")
	port       = flag.Int("port", 8333, "listen port")
)

func main() {
	flag.Parse()
	initLocalInfo()
	q := queue.New("channel")
	jRpc := rpc.NewJRpcServer(*port, q.Client())
	go jRpc.Listen()
	scanner := NewScanner(q.Client())
	if scanner != nil {
		scanner.Start()
	} else {
		log.Error("scanner start failed")
		return
	}
	select {}
}
