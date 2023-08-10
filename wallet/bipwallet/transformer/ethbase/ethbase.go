package ethbase

import (
	"fmt"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/system/address/eth"
	"github.com/33cn/chain33/system/crypto/secp256k1"
)

// ethBaseTransformer 转换基于以太坊地址规则的币种实现类
type ethBaseTransformer struct {
}

//PrivKeyToPub 32字节私钥生成公钥
func (e ethBaseTransformer) PrivKeyToPub(keyTy uint32, priv []byte) (pub []byte, err error) {
	if len(priv) != 32 {
		return nil, fmt.Errorf("invalid privkey")
	}
	//crypto.GetName(int(keyTy)
	//必须是secp256k1
	if crypto.GetName(int(keyTy)) != secp256k1.Name {
		panic("ethereum's crypto must secp256k1")
	}
	edcrypto, err := crypto.Load(secp256k1.Name, -1)
	if err != nil {
		return nil, err
	}

	edkey, err := edcrypto.PrivKeyFromBytes(priv)
	if err != nil {
		return nil, err
	}
	//压缩格式的公钥，在创建ETH地址的时候需要用到非压缩格式的公钥，即65字节，需要相应的转换
	compressPub := edkey.PubKey().Bytes()
	return compressPub, nil
}

func (e ethBaseTransformer) PubKeyToAddress(pub []byte) (add string, err error) {
	ethDriver, err := address.LoadDriver(eth.ID, -1)
	if err != nil {
		return "", err
	}

	return ethDriver.PubKeyToAddr(pub), nil

}
