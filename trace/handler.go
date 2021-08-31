package trace

import (
	"github.com/gorilla/mux"
	"github.com/multiformats/go-multiaddr"
	"net/http"
)
type peersResponse struct {
	Peers []Peer `json:"peers"`
}
// Peer holds information about a Peer.
type Peer struct {
	Address  string         `json:"address"`
	FullNode bool          `json:"fullNode"`
}

type chainStateResponse struct {
	//TODO
}
func (s *Service) blacklistPeersHandler(w http.ResponseWriter, r *http.Request) {
	var blackpeer []Peer
	//TODO 实现获取黑名单数据
	OK(w, peersResponse{
		Peers: blackpeer,
	})
}
func (s *Service) peersHandler(w http.ResponseWriter, r *http.Request) {
	var peers []Peer
	//TODO 获取当前连接的所有的节点
	OK(w, peersResponse{
		Peers: peers,
	})
}

func (s*Service)peerConnectHandler(w http.ResponseWriter, r *http.Request){
	//指定一个peer去发起连接
	addr, err := multiaddr.NewMultiaddr("/" + mux.Vars(r)["multi-address"])
	if err!=nil{
		BadRequest(w,err)
		return
	}
	log.Info("peerConnectHandler","addr",addr)
	//TODO 对指定的节点发起连接
	var peerAddr string
	OK(w,peerAddr)
}

// chainStateHandler returns the current chain state.
func (s *Service) chainStateHandler(w http.ResponseWriter, _ *http.Request) {

	OK(w, chainStateResponse{
	//TODO 获取区块链的状态数据
	})
}

//获取当前的拓扑结构数据
func (s*Service)topologyHandler(w http.ResponseWriter, _ *http.Request){
	//TODO DHT 路由表数据

}



