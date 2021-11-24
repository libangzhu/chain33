### grpc 简介
gRPC是一款语言中立、平台中立、开源的远程过程调用系统，即：gRPC 客户端和服务端可以在多种环境中运行交互，
例如用java 写一个服务端， 可以用go语言写客户端调用gRPC默认使用protocol buffers

***下面的接口使用说明在go环境下进行解释说明***

#### 接口约定
1. 导入grpc的Proto文件 chain33/types/p2p.pb.go
2. 创建grpc连接，gcon
3. 调用 p2p.pb.go 函数，对gcon 进行封装，NewP2PgserviceClient(gcon)

#### 交易接口
##### 1.1 构造并发送交易
要发送一个交易，均需要经过三个步骤：构造交易->交易签名->发送交易；

构造交易：填写交易的关键信息，从而构造一条完整的交易数据；
交易签名：对交易数据进行签名，即标识交易的所有者身份，也防止交易数据被篡改；
发送交易：将交易数据发送到区块链上去执行；

下面分别介绍三个接口

###### 1.1.1 构造交易 CreateRawTransaction
接口调用
```bigquery
 CreateRawTransaction(ctx context.Context, in *pb.CreateTx) (*pb.UnsignTx, error)

```
参数说明：
1. ctx context.Context 类型，默认可以用context.Background(),  其他用法可以参考context 包内容
2. CreateTx 结构体参数展示
```bigquery

type CreateTx struct {
	To          string `protobuf:"bytes,1,opt,name=to,proto3" json:"to,omitempty"`
	Amount      int64  `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
	Fee         int64  `protobuf:"varint,3,opt,name=fee,proto3" json:"fee,omitempty"`
	Note        []byte `protobuf:"bytes,4,opt,name=note,proto3" json:"note,omitempty"`
	IsWithdraw  bool   `protobuf:"varint,5,opt,name=isWithdraw,proto3" json:"isWithdraw,omitempty"`
	IsToken     bool   `protobuf:"varint,6,opt,name=isToken,proto3" json:"isToken,omitempty"`
	TokenSymbol string `protobuf:"bytes,7,opt,name=tokenSymbol,proto3" json:"tokenSymbol,omitempty"`
	ExecName    string `protobuf:"bytes,8,opt,name=execName,proto3" json:"execName,omitempty"`
	Execer      string `protobuf:"bytes,9,opt,name=execer,proto3" json:"execer,omitempty"`
}
```


|参数	|类型	|是否必填	|说明
|---- |----|----|----|
|To | string | 是 |发送到地址；如果是合约的充提则为合约地址，合约地址为execName转换而来，详情查看ConvertExectoAddr接口
|Amount	| int64	| 是	| 发送金额，注意基础货币单位为10^8
|Fee	|int64	|是	|手续费，注意基础货币单位为10^8,默认10^5,即Fee=1e5
|Note	|[]byte	|否	|备注,此处是字节数组, []byte("it's my tx note")
|IsToken	|bool	|否	|是否是token类型的转账 （非token转账这个不用填 包括平行链的基础代币转账也不用填）
|IsWithdraw	|bool	|是	|是否为从合约中提款的交易，普通转账为false
|TokenSymbol	|string	|否	|token 的 symbol （非token转账这个不用填）
|ExecName	|string	|否	|目标合约名，如果要构造平行链上的转账或普通转账，此参数置空
|Execer	|string	|是	|资产的执行器名称，如果是普通转账，此处应填coins，如果是构造平行链的基础代币，此处要填写user.p.xxx.coins token同上
返回参数
```bigquery
type UnsignTx struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}
```
|参数	|类型	|是否必填	|说明
|---- |----|----|----|
| Data|[]byte|是|待签名的已经构造完毕的交易

请求示例：
 假定 grpc服务地址：192.168.1.123:8802
 ```bigquery
 pb import"xxx/xxx/p2p.pb.go"
 gcon:= grpc.Dial(ctx,"192.168.1.123:8802")
 gclient:=pb.NewP2PgserviceClient(gcon)
 var txparam pb.CreateTx
 txparam.Amount=1e9
 txparam.Fee=1e5
 txparam.Execer="coins"
 txparam.To="16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
 unsignexTx,err:=gclient.CreateRawTransaction(context.Background(),&txparam)
 if err!=nil{
    //异常处理逻辑
 }
 
```
unsignexTx：返回的待签名的构造数据
