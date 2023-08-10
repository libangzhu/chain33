package ethbase

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/common"

	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/wallet/bipwallet/transformer"
)

//测试私钥生成公钥
var testprivkey = "c8729f05b10cc74d40feeb00376e11aa5b50e92b369d778b31b6e902c528f141"

func TestEthBaseTransformer_PrivKeyToPub(t *testing.T) {
	testPrivToPubToAddress(t, "ETH")
}

func testPrivToPubToAddress(t *testing.T, name string) {
	coinTrans, err := transformer.New(name)
	if err != nil {
		t.Errorf("new %s transformer error: %s", name, err)
	}
	pubByte, err := coinTrans.PrivKeyToPub(types.SECP256K1, common.FromHex(testprivkey))
	if err != nil {
		t.Errorf("%s PrivKeyToPub error: %s", name, err)
	}
	t.Log("pubsize:", len(pubByte), "pubstr:", common.Bytes2Hex(pubByte))
	addr, err := coinTrans.PubKeyToAddress(pubByte)
	assert.Nil(t, err)
	assert.Equal(t, "0xd83b69c56834e85e023b1738e69bfa2f0dd52905", addr)
	t.Log("ethaddress:", addr)

}
