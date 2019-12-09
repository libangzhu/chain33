package broadcast

import (
	"encoding/hex"
	"github.com/33cn/chain33/types"
)

func (s *Service) sendTx(tx *types.P2PTx, p2pData *types.BroadCastData, pid, peerAddr string) (doSend bool) {

	txHash := hex.EncodeToString(tx.Tx.Hash())
	ttl := tx.GetRoute().GetTTL()
	isLightSend := ttl >= s.node.GetP2pCfg().LightTxTTL
	//检测冗余发送
	ignoreSend := false

	//短哈希广播不记录发送过滤
	if !isLightSend {
		ignoreSend = addIgnoreSendPeerAtomic(s.txSendFilter, txHash, pid)
	}

	log.Debug("P2PSendTx", "txHash", txHash, "ttl", ttl, "isLightSend", isLightSend,
		"peerAddr", peerAddr, "ignoreSend", ignoreSend)

	if ignoreSend {
		return false
	}
	//超过最大的ttl, 不再发送
	if ttl > s.node.GetP2pCfg().MaxTTL {
		return false
	}

	//新版本且ttl达到设定值
	if isLightSend {
		p2pData.Value = &types.BroadCastData_LtTx{ //超过最大的ttl, 不再发送
			LtTx: &types.LightTx{
				TxHash: tx.Tx.Hash(),
				Route:  tx.GetRoute(),
			},
		}
	} else {
		p2pData.Value = &types.BroadCastData_Tx{Tx: tx}
	}
	return true
}

func (s *Service) recvTx(tx *types.P2PTx, pid, peerAddr string) {
	if tx.GetTx() == nil {
		return
	}
	txHash := hex.EncodeToString(tx.GetTx().Hash())
	//将节点id添加到发送过滤, 避免冗余发送
	addIgnoreSendPeerAtomic(s.txSendFilter, txHash, pid)
	//重复接收
	isDuplicate := checkAndRegFilterAtomic(s.txFilter, txHash)
	log.Debug("recvTx", "tx", txHash, "ttl", tx.GetRoute().GetTTL(), "peerAddr", peerAddr, "duplicateTx", isDuplicate)
	if isDuplicate {
		return
	}
	//有可能收到老版本的交易路由,此时route是空指针
	if tx.GetRoute() == nil {
		tx.Route = &types.P2PRoute{TTL: 1}
	}
	s.txFilter.Add(txHash, tx.GetRoute())

	msg := s.client.NewMessage("mempool", types.EventTx, tx.GetTx())
	errs := s.client.Send(msg, false)
	if errs != nil {
		log.Error("recvTx", "process EventTx msg Error", errs.Error())
	}

}

func (s *Service) recvLtTx(tx *types.LightTx, pid, peerAddr string) {

	txHash := hex.EncodeToString(tx.TxHash)
	//将节点id添加到发送过滤, 避免冗余发送
	addIgnoreSendPeerAtomic(s.txSendFilter, txHash, pid)
	exist := s.txFilter.QueryRecvData(txHash)
	log.Debug("recvLtTx", "txHash", txHash, "ttl", tx.GetRoute().GetTTL(), "peerAddr", peerAddr, "exist", exist)
	//本地不存在, 需要向对端节点发起完整交易请求. 如果存在则表示本地已经接收过此交易, 不做任何操作
	if !exist {

		query := &types.P2PQueryData{}
		query.Value = &types.P2PQueryData_TxReq{
			TxReq: &types.P2PTxReq{
				TxHash: tx.TxHash,
			},
		}
		//发布到指定的节点
		s.queryStream(pid, query)
	}
}
