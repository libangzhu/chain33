// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package account 实现chain33 区块链资产操作
package account

//package for account manger
//1. load from db
//2. save to db
//3. KVSet
//4. Transfer
//5. Add
//6. Sub
//7. Account balance query
//8. gen a private key -> private key to address (bitcoin likes)

import (
	"fmt"
	"strings"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
)

var alog = log.New("module", "account")

// DB for account
type DB struct {
	db                   dbm.KV
	accountKeyPerfix     []byte
	execAccountKeyPerfix []byte
	execer               string
	symbol               string
}

// NewCoinsAccount 新建账户
func NewCoinsAccount() *DB {
	prefix := "mavl-coins-bty-"
	return newAccountDB(prefix)
}

// NewAccountDB 新建DB账户
func NewAccountDB(execer string, symbol string, db dbm.KV) (*DB, error) {
	//如果execer 和  symbol 中存在 "-", 那么创建失败
	if strings.ContainsRune(execer, '-') {
		return nil, types.ErrExecNameNotAllow
	}
	if strings.ContainsRune(symbol, '-') {
		return nil, types.ErrSymbolNameNotAllow
	}
	accDB := newAccountDB(symbolPrefix(execer, symbol))
	accDB.execer = execer
	accDB.symbol = symbol
	accDB.SetDB(db)
	return accDB, nil
}

func newAccountDB(prefix string) *DB {
	acc := &DB{}
	acc.accountKeyPerfix = []byte(prefix)
	acc.execAccountKeyPerfix = append([]byte(prefix), []byte("exec-")...)
	//alog.Warn("NewAccountDB", "prefix", prefix, "key1", string(acc.accountKeyPerfix), "key2", string(acc.execAccountKeyPerfix))
	return acc
}

// SetDB set db
func (acc *DB) SetDB(db dbm.KV) *DB {
	acc.db = db
	return acc
}

// LoadAccount 根据地址载入账户
func (acc *DB) LoadAccount(addr string) *types.Account {
	value, err := acc.db.Get(acc.AccountKey(addr))
	if err != nil {
		return &types.Account{Addr: addr}
	}
	var acc1 types.Account
	err = types.Decode(value, &acc1)
	if err != nil {
		panic(err) //数据库已经损坏
	}
	return &acc1
}

// CheckTransfer 检查交易
func (acc *DB) CheckTransfer(from, to string, amount int64) error {
	if !types.CheckAmount(amount) {
		return types.ErrAmount
	}
	accFrom := acc.LoadAccount(from)
	b := accFrom.GetBalance() - amount
	if b < 0 {
		return types.ErrNoBalance
	}
	return nil
}

// Transfer 执行交易
func (acc *DB) Transfer(from, to string, amount int64) (*types.Receipt, error) {
	if !types.CheckAmount(amount) {
		return nil, types.ErrAmount
	}
	accFrom := acc.LoadAccount(from)
	accTo := acc.LoadAccount(to)
	if accFrom.Addr == accTo.Addr {
		return nil, types.ErrSendSameToRecv
	}
	if accFrom.GetBalance()-amount >= 0 {
		copyfrom := *accFrom
		copyto := *accTo

		accFrom.Balance = accFrom.GetBalance() - amount
		accTo.Balance = accTo.GetBalance() + amount

		receiptBalanceFrom := &types.ReceiptAccountTransfer{
			Prev:    &copyfrom,
			Current: accFrom,
		}
		receiptBalanceTo := &types.ReceiptAccountTransfer{
			Prev:    &copyto,
			Current: accTo,
		}

		acc.SaveAccount(accFrom)
		acc.SaveAccount(accTo)
		return acc.transferReceipt(accFrom, accTo, receiptBalanceFrom, receiptBalanceTo), nil
	}

	return nil, types.ErrNoBalance
}

func (acc *DB) depositBalance(execaddr string, amount int64) (*types.Receipt, error) {
	if !types.CheckAmount(amount) {
		return nil, types.ErrAmount
	}
	acc1 := acc.LoadAccount(execaddr)
	copyacc := *acc1
	acc1.Balance += amount
	receiptBalance := &types.ReceiptAccountTransfer{
		Prev:    &copyacc,
		Current: acc1,
	}
	acc.SaveAccount(acc1)
	ty := int32(types.TyLogDeposit)
	ty = types.TyLogDeposit
	log1 := &types.ReceiptLog{
		Ty:  ty,
		Log: types.Encode(receiptBalance),
	}
	kv := acc.GetKVSet(acc1)
	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kv,
		Logs: []*types.ReceiptLog{log1},
	}, nil
}

func (acc *DB) transferReceipt(accFrom, accTo *types.Account, receiptFrom, receiptTo proto.Message) *types.Receipt {
	ty := int32(types.TyLogTransfer)
	log1 := &types.ReceiptLog{
		Ty:  ty,
		Log: types.Encode(receiptFrom),
	}
	log2 := &types.ReceiptLog{
		Ty:  ty,
		Log: types.Encode(receiptTo),
	}
	kv := acc.GetKVSet(accFrom)
	kv = append(kv, acc.GetKVSet(accTo)...)
	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kv,
		Logs: []*types.ReceiptLog{log1, log2},
	}
}

// SaveAccount 保存账户到数据库
func (acc *DB) SaveAccount(acc1 *types.Account) {
	set := acc.GetKVSet(acc1)
	for i := 0; i < len(set); i++ {
		acc.db.Set(set[i].GetKey(), set[i].Value)
	}
}

// GetKVSet 将账户数据转为数据库存储kv
func (acc *DB) GetKVSet(acc1 *types.Account) (kvset []*types.KeyValue) {
	value := types.Encode(acc1)
	kvset = append(kvset, &types.KeyValue{
		Key:   acc.AccountKey(acc1.Addr),
		Value: value,
	})
	return kvset
}

// LoadAccounts 从stateDB中载入若干账户
// TODO:使用API的方式访问,暂时与LoadAccounts()共存,后续将删除LoadAccounts()
func (acc *DB) LoadAccounts(api client.QueueProtocolAPI, addrs []string) (accs []*types.Account, err error) {
	header, err := api.GetLastHeader()
	if err != nil {
		return nil, err
	}
	return acc.loadAccountsHistory(api, addrs, header.GetStateHash())
}

// LoadAccountsDB 载入账户
func (acc *DB) LoadAccountsDB(addrs []string) (accs []*types.Account, err error) {
	for i := 0; i < len(addrs); i++ {
		acc1 := acc.LoadAccount(addrs[i])
		accs = append(accs, acc1)
	}
	return accs, nil
}

// AccountKey return the key of address in DB
func (acc *DB) AccountKey(address string) (key []byte) {
	key = append(key, acc.accountKeyPerfix...)
	key = append(key, []byte(address)...)
	return key
}

func symbolPrefix(execer string, symbol string) string {
	return fmt.Sprintf("mavl-%s-%s-", execer, symbol)
}

func symbolExecPrefix(execer string, symbol string) string {
	return fmt.Sprintf("mavl-%s-%s-exec", execer, symbol)
}

// GetTotalCoins 获取代币总量
func (acc *DB) GetTotalCoins(api client.QueueProtocolAPI, in *types.ReqGetTotalCoins) (reply *types.ReplyGetTotalCoins, err error) {
	req := types.IterateRangeByStateHash{}
	req.StateHash = in.StateHash
	req.Count = in.Count
	start := symbolPrefix(in.Execer, in.Symbol)
	end := symbolExecPrefix(in.Execer, in.Symbol)
	if in.StartKey == nil {
		req.Start = []byte(start)
	} else {
		req.Start = in.StartKey
	}
	req.End = []byte(end)
	return api.StoreGetTotalCoins(&req)
}

func (acc *DB) loadAccountsHistory(api client.QueueProtocolAPI, addrs []string, stateHash []byte) (accs []*types.Account, err error) {
	get := types.StoreGet{StateHash: stateHash}
	for i := 0; i < len(addrs); i++ {
		get.Keys = append(get.Keys, acc.AccountKey(addrs[i]))
	}

	values, err := api.StoreGet(&get)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(values.Values); i++ {
		value := values.Values[i]
		if value == nil {
			accs = append(accs, &types.Account{Addr: addrs[i]})
		} else {
			var acc types.Account
			err := types.Decode(value, &acc)
			if err != nil {
				return nil, err
			}
			accs = append(accs, &acc)
		}
	}
	return accs, nil
}

// GetBalance 获取某个状态下账户余额
func (acc *DB) GetBalance(api client.QueueProtocolAPI, in *types.ReqBalance) ([]*types.Account, error) {
	switch in.GetExecer() {
	case types.ExecName("coins"):
		addrs := in.GetAddresses()
		var exaddrs []string
		for _, addr := range addrs {
			if err := address.CheckAddress(addr); err != nil {
				addr = address.ExecAddress(addr)
			}
			exaddrs = append(exaddrs, addr)
		}
		var accounts []*types.Account
		var err error
		if len(in.StateHash) == 0 {
			accounts, err = acc.LoadAccounts(api, exaddrs)
		} else {
			hash, err := common.FromHex(in.StateHash)
			if err != nil {
				return nil, err
			}
			accounts, err = acc.loadAccountsHistory(api, exaddrs, hash)
		}
		if err != nil {
			log.Error("GetBalance", "err", err.Error())
			return nil, err
		}
		return accounts, nil
	default:
		execaddress := address.ExecAddress(in.GetExecer())
		addrs := in.GetAddresses()
		var accounts []*types.Account
		for _, addr := range addrs {
			var account *types.Account
			var err error
			if len(in.StateHash) == 0 {
				account, err = acc.LoadExecAccountQueue(api, addr, execaddress)
			} else {
				hash, err := common.FromHex(in.StateHash)
				if err != nil {
					return nil, err
				}
				account, err = acc.LoadExecAccountHistoryQueue(api, addr, execaddress, hash)
			}
			if err != nil {
				log.Error("GetBalance", "err", err.Error())
				continue
			}
			accounts = append(accounts, account)
		}
		return accounts, nil
	}
}
