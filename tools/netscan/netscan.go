package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/system/p2p/dht"
	"github.com/33cn/chain33/system/p2p/dht/manage"
	dprotol "github.com/33cn/chain33/system/p2p/dht/protocol"
	p2pty "github.com/33cn/chain33/system/p2p/dht/types"
	"github.com/33cn/chain33/tools/netscan/rpc"
	"github.com/33cn/chain33/types"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	peerInfoProtoOld protocol.ID = "/chain33/peerinfoReq/1.0.0"
	peerInfoProto                = "/chain33/peer-info/1.0.0"
	cfgPath                      = flag.String("f", "scan.toml", "config file")
)

type NetScan struct {
	host          core.Host
	discovery     *dht.Discovery
	peerInfoManag *manage.PeerInfoManager
	cancel        context.CancelFunc
	ctx           context.Context
	seeds         []string
	cli           queue.Client
}

func NewScanner(client queue.Client) *NetScan {
	log.Info("NewScanner", "cfg path", *cfgPath)
	ctx, cancel := context.WithCancel(context.Background())
	cfg := types.NewChain33Config(types.ReadFile(*cfgPath))
	mcfg := &p2pty.P2PSubConfig{}
	types.MustDecode(cfg.GetSubConfig().P2P[p2pty.DHTTypeName], mcfg)
	srcMAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", mcfg.Port))
	if err != nil {
		log.Error("Unable to construct multiaddr %v", err)
		return nil
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrs(srcMAddr),
		libp2p.ConnectionManager(connmgr.NewConnManager(2000, 20000, time.Minute*2)),
	}

	book := dht.NewAddrBook(cfg.GetModuleConfig().P2P)
	priv := book.Randkey()
	opts = append(opts, libp2p.Identity(priv))
	h, err := libp2p.New(
		ctx,
		opts...,
	)
	if err != nil {
		log.Error("new p2p", "err", err)
		panic(err)
	}
	dis := dht.InitDhtDiscovery(ctx, h, nil, cfg, mcfg)
	peers := ConvertPeers(mcfg.Seeds)
	log.Info("netscan", "hostId", h.ID(), "seeds", mcfg.Seeds)
	var seedpid []string
	for pid := range peers {
		seedpid = append(seedpid, pid)
	}

	return &NetScan{host: h,
		discovery:     dis,
		peerInfoManag: manage.NewPeerInfoManager(ctx, h, nil, time.Hour),
		cancel:        cancel,
		ctx:           ctx,
		seeds:         seedpid,
		cli:           client}
}
func (n *NetScan) Start() {
	n.discovery.Start()
	go n.ScanNetPeerInfos()
	go n.ScanNetPeers()
	go n.TicketWrite()
	go n.GetClosestPeers()
	//queue msg
	go n.subMsg()
}

func (n *NetScan) subMsg() {
	for {
		n.cli.Sub("p2p")
		for msg := range n.cli.Recv() {
			switch msg.Ty {
			case rpc.EventPeerLocaltionInfo:
				//ReflushLocalInfo(*gossipPath)
				infos := GetpeerLocaltionInfo()
				log.Info("subMsg", "infos", len(infos))
				msg.Reply(n.cli.NewMessage("rpc", rpc.EventPeerLocaltionInfo, &rpc.ConriesInfo{Countries: infos}))
			}
		}

	}
}
func (n *NetScan) Close() {
	n.discovery.Close()
	n.cancel()
}

func (n *NetScan) GetClosestPeers() {
	for {
		time.Sleep(time.Second * 20)
		if len(n.host.Peerstore().Peers()) == 0 {
			continue
		}

		var wg sync.WaitGroup
		for _, peer := range n.host.Peerstore().Peers() {
			wg.Add(1)
			go n.connectClosesPeers(peer, &wg)
		}
		wg.Wait()

	}
}
func (n *NetScan) connectClosesPeers(peer peer.ID, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(n.ctx, time.Second*10)
	defer cancel()
	peerChan, err := n.discovery.GetDht().GetClosestPeers(ctx, peer.String())
	if err != nil {
		return
	}
	for {
		p, ok := <-peerChan
		if !ok {
			return
		}
		n.host.Network().DialPeer(n.ctx, p)

	}

}
func (n *NetScan) fetchConnToPeer(peerID core.PeerID, wg *sync.WaitGroup) {
	defer wg.Done()

	peerchan, err := n.discovery.FindPeersConnectedToPeer(peerID)
	if err != nil {
		log.Error("FetchConnToPeer", "err", err)
		return
	}

	for {
		pinfo, ok := <-peerchan
		if !ok {
			return
		}
		err = n.host.Connect(n.ctx, *pinfo)
		log.Error("fetchConnToPeer", "err", err)
	}

}

//获取连接的节点当前的信息
func (n *NetScan) ScanNetPeerInfos() { //
	var wg sync.WaitGroup
	for {
		time.Sleep(time.Second * 5)
		if len(n.host.Network().Conns()) == 0 {
			time.Sleep(time.Second)
			continue
		}

		for _, con := range n.host.Network().Conns() {
			if con.RemotePeer() == n.host.ID() {
				continue
			}

			n.getPeerInfo(con.RemotePeer(), nil)
			wg.Add(1)
			go n.fetchConnToPeer(con.RemotePeer(), &wg)
		}
		wg.Wait()
		log.Info("ScanNetPeerInfos", "当前连接的节点数量： num.+++++", len(n.host.Network().Peers()), "当前的连接数", len(n.host.Network().Conns()))
		pinfos := n.peerInfoManag.FetchAll()
		log.Info("scanNetPeer", "peerInfoManag size", len(pinfos))

	}

}

func (n *NetScan) TicketWrite() {
	ticker := time.NewTicker(time.Minute * 5)

	for {
		<-ticker.C
		var versionM = make(map[string]map[string]int64)

		serviceF := Createfile("onservice") //
		unsyncF := Createfile("unsync")
		versionF := Createfile("version")
		allNodeF := Createfile("onlinepids")
		localionsF := Createfile("localtions")
		countryF := Createfile("countryinfos")
		var standerHeight int64
		for _, seed := range n.seeds {
			peer := n.peerInfoManag.Fetch(peer.ID(seed))
			seedHeight := peer.GetHeader().GetHeight()
			if standerHeight < seedHeight {
				standerHeight = seedHeight
			}
		}
		pinfos := n.peerInfoManag.FetchAll()
		for _, info := range pinfos {
			if info.Version == "" {
				info.Version = "未识别版本"
			}
			if v, ok := versionM[info.Version]; ok {
				v[info.Name] = info.Header.Height
			} else {

				versionM[info.Version] = make(map[string]int64)
				versionM[info.Version][info.Name] = info.Header.Height
			}
			fmt.Println("peerHeight","height",info.Header.GetHeight(),"standerHeight",standerHeight)
			if info.Header.Height+512 >= standerHeight { //512个以内，被认为是同步的
				//增加版本号
				serviceF.WriteString(fmt.Sprintf("%v@%v@%v\n", info.Name, fmt.Sprintf("%s:%d", info.Addr, info.Port), info.Version))
			} else {

				unsyncF.WriteString(fmt.Sprintf("%v@%v@diff:%d@%v\n", info.Name,
					fmt.Sprintf("%s:%d", info.Addr, info.Port), standerHeight-info.Header.Height, info.Version))
			}

		}
		for ver, vV := range versionM {
			versionF.WriteString(fmt.Sprintf("%v-----%d\n", ver, len(vV)))
		}
		var tempLocalInfo = NewLocationInfo()
		var countryData = make(map[string][]string)

		for _, peer := range n.host.Peerstore().Peers() {
			pinfo := n.host.Peerstore().PeerInfo(peer)
			//fliter pubpi ip
			var ip string
			for _, addr := range pinfo.Addrs {
				addrsplites := strings.Split(addr.String(), "/")
				if len(addrsplites) >= 3 {
					//log.Info("isPublicIP","ippp-----------------------",addrsplites[2])
					if isPublicIP(addrsplites[2]) {
						ip = addrsplites[2]
						break
					}
				}
			}
			ipdata := n.CheckIp(ip)
			if ipdata != nil {
				ipdata.City = "city"
				ipdata.Region = "region"
				tempLocalInfo.Add(ipdata, pinfo.ID.Pretty())
			}

			allNodeF.WriteString(fmt.Sprintf("%s@%v\n", pinfo.ID, ip))
		}
		//
		tempLocalInfo.mtx.Lock()
		countryinfo := fmt.Sprintf("\n+++++++++++++++++++Country AND  Region Num:", len(tempLocalInfo.locationMap))
		for country, regionMap := range tempLocalInfo.locationMap {
			var pidNum int
			var pids []string
			for region, citymap := range regionMap {
				for city, info := range citymap {
					var cityinfo string
					pidNum += len(info.pids)
					//pids = append(pids, info.pids...)
					for pid := range info.pids {
						pids = append(pids, pid)
					}
					if region == city {
						city = ""
					}
					if country == region {
						region = ""
					}
					cityinfo += "\n" + country + region + city + fmt.Sprintf("node num:%v", len(info.pids))
					log.Info(cityinfo)
				}
			}
			countryData[country] = pids
			countryinfo += "\n" + country + "node num:" + fmt.Sprintf("%v", pidNum)
		}
		tempLocalInfo.mtx.Unlock()
		if ReflushLocalInfo(*gossipPath, tempLocalInfo) {
			//log.Info("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
			locaInfo = tempLocalInfo
		}

		jbytes, _ := json.Marshal(countryData)
		tempLocalInfo.Stat = true
		localionsF.WriteString(countryinfo)
		countryF.WriteString(string(jbytes))

		versionF.Close()
		serviceF.Close()
		unsyncF.Close()
		allNodeF.Close()
		countryF.Close()
		localionsF.Close()

	}
}

func (n *NetScan) CheckIp(ip string) *IP {
	var ipdata = new(IP)
	if v, _ := KvDb.Get([]byte(ip)); v != nil {
		json.Unmarshal(v, ipdata)
	} else {
		ipinfo := LocationAPI(ip)
		if ipinfo == nil || ipinfo.Code != 0 {
			return nil
		}
		jbtes, _ := json.Marshal(ipinfo.Data)
		json.Unmarshal(jbtes, ipdata)
		KvDb.Set([]byte(ip), jbtes)
	}
	return ipdata

}

//通过peerstore 去不断尝试连接网络中的节点
func (n *NetScan) ScanNetPeers() {
	for {
		<-time.After(time.Second * 20)
		var wg sync.WaitGroup
		for _, peer := range n.host.Peerstore().Peers() {
			if peer == n.host.ID() {
				continue
			}
			err := n.host.Connect(n.ctx, n.host.Peerstore().PeerInfo(peer))
			if err != nil {
				continue
			}
			//var wg sync.WaitGroup
			wg.Add(1)
			go n.fetchConnToPeer(peer, &wg)

		}
		wg.Wait()
		log.Info("ScanNetPeers", "peerstore 数量>>>>>", len(n.host.Peerstore().Peers()), "conns", len(n.host.Network().Conns()))

	}

}

func (n *NetScan) getPeerInfo(peerID core.PeerID, stream network.Stream) {
	msgReq := &types.MessagePeerInfoReq{}
	var resp types.MessagePeerInfoResp
	var err error
	var reNum int
ReConn:
	if stream == nil {
		//protocol.ID(peerInfoProto)
		ctx, cancel := context.WithTimeout(n.ctx, time.Second*5)
		defer cancel()
		stream, err = n.host.NewStream(ctx, peerID, peerInfoProtoOld, protocol.ID(peerInfoProto))
		if err != nil {
			//	log.Error("getPeerInfo", "NewStreamErr", err, "peerID", peerID)
			return
		}

	}

	defer dprotol.CloseStream(stream)
	err = dprotol.WriteScanStream(msgReq, stream)
	if err != nil {
		if err.Error() == "stream reset" {
			time.Sleep(time.Second)
			if reNum < 10 {
				reNum++
				goto ReConn
			}
		}
		return
	}
	err = dprotol.ReadSscanStream(&resp, stream)
	if err != nil {
		if err.Error() == "stream reset" {
			if reNum < 10 {
				reNum++
				goto ReConn
			}
		}
		return
	}
	peerinfo := resp.GetMessage()
	n.peerInfoManag.Refresh(&types.Peer{Name: peerinfo.Name, Addr: peerinfo.Addr, Port: peerinfo.GetPort(), MempoolSize: peerinfo.GetMempoolSize(),
		Header: peerinfo.GetHeader(), Version: peerinfo.GetVersion(), LocalDBVersion: peerinfo.GetLocalDBVersion(), StoreDBVersion: peerinfo.GetStoreDBVersion(),
		Self: false,
	})

	//stream.Close()

}

func ConvertPeers(peers []string) map[string]*peer.AddrInfo {
	pinfos := make(map[string]*peer.AddrInfo, len(peers))
	for _, addr := range peers {
		addrm, _ := multiaddr.NewMultiaddr(addr)
		peerinfo, err := peer.AddrInfoFromP2pAddr(addrm)
		if err != nil {
			continue
		}
		pinfos[peerinfo.ID.String()] = peerinfo
	}
	return pinfos

}

/*
tcp/ip协议中，专门保留了三个IP地址区域作为私有地址，其地址范围如下：
10.0.0.0/8：10.0.0.0～10.255.255.255
172.16.0.0/12：172.16.0.0～172.31.255.255
192.168.0.0/16：192.168.0.0～192.168.255.255
*/
func isPublicIP(ip string) bool {
	IP := net.ParseIP(ip)
	if IP == nil || IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}
