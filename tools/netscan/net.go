package main

import (
	"errors"
	"fmt"
	"github.com/33cn/chain33/types"
	"math/big"
	"strings"
	"time"
)

//RateCalculate
func (n *NetScan) RateCalculate(ratebytes float64) string {
	kbytes := ratebytes / 1024
	rate := fmt.Sprintf("%.3f KB/s", kbytes)

	if kbytes/1024 > 0.1 {
		rate = fmt.Sprintf("%.3f MB/s", kbytes/1024)
	}
	return rate
}

//ParaseRate
func (n *NetScan) ParaseRate(rate string) (float64, error) {
	kb := "KB/s"
	mb := "MB/s"
	if strings.HasSuffix(rate, kb) {
		kbytesStr := strings.TrimSuffix(rate, kb)
		bf := big.NewFloat(1)
		fmt.Println("kttes:", strings.TrimSpace(kbytesStr))
		bf, _ = bf.SetString(strings.TrimSpace(kbytesStr))
		kbs, _ := bf.Float64()
		return kbs * 1024, nil
	} else if strings.HasSuffix(rate, mb) {
		mbytesStr := strings.TrimSuffix(rate, mb)
		bf := big.NewFloat(1)
		fmt.Println("kttes:", strings.TrimSpace(mbytesStr))
		bf, _ = bf.SetString(strings.TrimSpace(mbytesStr))
		kbs, _ := bf.Float64()
		return kbs * 1024 * 1024, nil
	}

	return 0, errors.New("rate format error")
}

func (n *NetScan) ParaseRuntime(rate string) (float64, error) {
	//mt:="min"
	//ht:="hour"
	//st:="second"
	//if strings.HasSuffix(rate ,mt){
	//	kbytesStr:= strings.TrimSuffix(rate,mt)
	//
	//}
	//
	return 0, nil

}

func (n *NetScan) statisticNetinfo() {
	netInfoF := Createfile("netinfo")
	var writeStr string
	//统计平均网路带宽，细分为avage RateIn ,Rateout
	var totalOutBounds int32
	var totalInBounds int32
	var totalRatein float64
	var totalRateout float64
	var totalRoutingtable int32
	var peerNum int32
	n.netInfo.Range(func(key, value interface{}) bool {
		//netInfoF.WriteString(fmt.Sprintf("%v@"),key.(string))
		peerNum++
		info := value.(*types.NodeNetInfo)
		totalOutBounds += info.Outbounds
		totalInBounds += info.Inbounds
		n.UpdateConnectInBounds(info.Inbounds)
		n.UpdateConnectOutBounds(info.Outbounds)

		totalRoutingtable += info.GetRoutingtable()
		//统计带宽
		rateInBytes, err := n.ParaseRate(info.Ratein)
		if err != nil {
			log.Error("")
		}
		rateoutBytes, err := n.ParaseRate(info.Rateout)
		n.StatisticNetRate(int32(rateoutBytes + rateInBytes))
		totalRatein += rateInBytes
		totalRateout += rateoutBytes
		writeStr += fmt.Sprintf("%v@inbounds:%d@outbounds:%d@ratein:%v@rateout:%v@routingTable:%v\n", key, info.Inbounds, info.Outbounds, info.Ratein, info.Rateout, info.GetRoutingtable())
		return true
	})
	//平均连接数
	averageOutBounds := totalOutBounds / peerNum
	averageInBounds := totalInBounds / peerNum
	averageRatein := totalRatein / float64(peerNum)
	avaerageRateOut := totalRateout / float64(peerNum)
	avaerageRate := (totalRatein + totalRateout) / float64(peerNum)

	netInfoF.WriteString(writeStr)
	netInfoF.WriteString("--------------\n")
	netInfoF.WriteString(fmt.Sprintf("平均连出总数:%v\n平均连入节点数:%v\n", averageOutBounds, averageInBounds))
	netInfoF.WriteString(fmt.Sprintf("平均下行带宽:%v\n平均上行带宽:%v\n", n.RateCalculate(averageRatein), n.RateCalculate(avaerageRateOut)))
	netInfoF.WriteString(fmt.Sprintf("平均带宽:%v\n", n.RateCalculate(avaerageRate)))
	netInfoF.WriteString(fmt.Sprintf("平均的路由表数量:%v\n", totalRoutingtable/peerNum))

	//刷新到数据库
	now := fmt.Sprintf("%v-%v", time.Now().Format("2006-01-02"), time.Now().Hour())
	//key=level-netrate-2022-03-23-15
	log.Info("netrate save", "key:", NetRatePrefix+now, "value:", n.RateCalculate(avaerageRate))
	KvDb.Set([]byte(NetRatePrefix+now), []byte(n.RateCalculate(avaerageRate)))
	netInfoF.Close()
	//write ConnectInBounds
	/*
	inF := Createfile("inbouds")
	outF := Createfile("outbouds")
	netRateF := Createfile("netrate")
	n.InConnNum.Range(func(key, value interface{}) bool {
		inF.WriteString(fmt.Sprintf("%v,%v\n", key, value))
		return true
	})

	n.OutConnNum.Range(func(key, value interface{}) bool {
		outF.WriteString(fmt.Sprintf("%v,%v\n", key, value))
		return true
	})

	n.NetRate.Range(func(key, value interface{}) bool {
		netRateF.WriteString(fmt.Sprintf("%v,%v\n", key.(int32)/1000, value))
		return true
	})

	inF.Close()
	outF.Close()
	netRateF.Close()

	 */
}

func (n *NetScan) UpdateConnectInBounds(bounds int32) {
	if v, ok := n.InConnNum.Load(bounds); ok {
		count := v.(int32) + 1
		n.InConnNum.Store(bounds, count)
		return
	}
	var count int32 = 1
	n.InConnNum.Store(bounds, count)
}

func (n *NetScan) UpdateConnectOutBounds(bounds int32) {
	if v, ok := n.OutConnNum.Load(bounds); ok {
		count := v.(int32) + 1
		n.OutConnNum.Store(bounds, count)
		return
	}
	var count int32 = 1
	n.OutConnNum.Store(bounds, count)
}

func (n *NetScan) StatisticNetRate(ratebytes int32) {

	if v, ok := n.NetRate.Load(ratebytes); ok {
		count := v.(int32) + 1
		n.NetRate.Store(ratebytes, count)
		return
	}

	var count int32 = 1
	n.NetRate.Store(ratebytes, count)
}
