package ethbase

import "github.com/33cn/chain33/wallet/bipwallet/transformer"

func init() {
	//注册
	transformer.Register("ETH", &ethBaseTransformer{})
}
