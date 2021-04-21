package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	//"code.aliyun.com/chain33/netscan/nodeaward/account"
	//"code.aliyun.com/chain33/netscan/nodeaward/common"
	//"code.aliyun.com/chain33/netscan/common"
	//"code.aliyun.com/chain33/netscan/types"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	//cty	"gitlab.33.cn/chain33/chain33/system/dapp/coins/types"
	//"gitlab.33.cn/chain33/chain33/common/crypto"
	//"gitlab.33.cn/chain33/chain33/types"
)

var (
	hexkey    = ""

)

func SendCoins(to string, amount int64) error {
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		log.Info(err.Error())
		return err
	}

	hexbytes, err := common.FromHex(hexkey)
	if err != nil {
		log.Info(err.Error())
		return err
	}

	priv, err := cr.PrivKeyFromBytes(hexbytes)
	if err != nil {
		log.Info(err.Error())
		return err
	}

	addrfrom := address.PubKeyToAddress(priv.PubKey().Bytes())
	addrto := to
	log.Info("SendCoins", "addrfrom", addrfrom)
	log.Info("SendCoins", "to", addrto)
	log.Info("SendCoins", "amount", amount)

	v := &cty.CoinsAction_Transfer{&types.AssetsTransfer{Cointoken: "BTY", Amount: amount, Note: []byte("node award")}}
	transfer := &cty.CoinsAction{Value: v, Ty: cty.CoinsActionTransfer}
	//var nonce int64 = 10240000
	tx := &types.Transaction{Execer: []byte("coins"), Payload: types.Encode(transfer), Fee: 1e5, To: addrto, Nonce: rand.New(rand.NewSource(time.Now().UnixNano())).Int63()}
	//tx.SetExpire(time.Second * 120)
	tx.Sign(types.SECP256K1, priv)

	poststr := fmt.Sprintf(`{"jsonrpc":"2.0","id":2,"method":"Chain33.SendTransaction","params":[{"data":"%v"}]}`,
		common.ToHex(types.Encode(tx)))
	var urls []string = []string{"http://183.129.226.77:8801", "http://116.63.171.186:8801"}
	var senderr error
	var txid string
	for _, url := range urls {
		resp, err := http.Post(url, "application/json", bytes.NewBufferString(poststr))
		if err != nil {
			senderr = err
			continue
		}
		defer resp.Body.Close()
		fmt.Println(resp.Header)
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Info("err:", err.Error())
			senderr = err
			continue
		}

		fmt.Printf("returned JSON: %s\n", string(b))
		var jdata struct {
			Id     int32       `json:"id"`
			Result interface{} `json:"result"`
			Error  interface{} `json:"error"`
		}
		json.Unmarshal(b, &jdata)
		if jdata.Error != nil {
			senderr = fmt.Errorf(jdata.Error.(string))
			continue
		}

		txid = jdata.Result.(string)
		break
	}
	log.Info("send", "txid", txid)
	if len(txid) != 0 {
		return nil
	}
	return senderr
}
