package main

import (
	"fmt"
	"strconv"
	"time"
)

const (//2800个空投coin
	OnServiceCoins = 1250*2
	OnlineCoins    = 150*2
)

func normalAward(bestfilepath []string) {

	var totalSend float32
	var onlineChans = make(schan, 128)
	var onserviceChans = make(schan, 128)
	//扫描文件夹下的文件
	var txcount int = 0
	var scanfs ScanFs

	for _, path := range bestfilepath {
		scanfs.ScanFolderFiles(path, onlineChans, onserviceChans)
	}
	close(onlineChans)
	close(onserviceChans)
	//onlineChans 所有全网节点
	//onserviceChans 所有能对外提供服务的节点
	var OnseviceMap map[string]string
	//var OnServiceIpMap map[string][]string
	var OnServiceIpMap map[string]map[string]interface{}
	var read ReadParase
	var filebytes [][]byte
	for filename := range onserviceChans {
		log.Info("normalAward", "filename", filename)
		rbs := read.ReadFile(filename)
		filebytes = append(filebytes, rbs)
	}
	OnseviceMap, OnServiceIpMap = read.parseOnServFileContentMap(filebytes)
	log.Info("normalAward", "onservieIpMap", len(OnServiceIpMap))
	var onserviceCount int
	var noserviceCount int
	var onlinebs [][]byte
	for filename := range onlineChans {
		log.Info("normalAward", "onlineChans filename", filename)
		rbs := read.ReadFile(filename)
		onlinebs = append(onlinebs, rbs)
	}
	toaddrs := read.parseOnlineFileAddr(onlinebs)
	log.Info("total nodes", "nodes num", len(toaddrs))

	for  addr := range toaddrs {
		if _, ok := OnseviceMap[addr]; ok {
			//提供服务的几点
			onserviceCount++
			continue
		}
		noserviceCount++
	}
	//2500+300=2800

	sendOnServiceCoinsStr := fmt.Sprintf("%.3f", (OnServiceCoins)/float32(len(OnServiceIpMap)))
	sendNoSerViceCoinsStr := fmt.Sprintf("%.3f", (OnlineCoins)/float32(noserviceCount+len(OnServiceIpMap)))
	sendOnServiceCoins, _ := strconv.ParseFloat(sendOnServiceCoinsStr, 64)
	sendNoSerViceCoins, _ := strconv.ParseFloat(sendNoSerViceCoinsStr, 64)

	log.Info("onserviceCount", "count", onserviceCount,"OnseviceMap",len(OnseviceMap), "average", sendOnServiceCoins)
	log.Info("allCount", "count", noserviceCount+len(OnServiceIpMap), "average", sendNoSerViceCoins)
	if noserviceCount+len(OnServiceIpMap) < 300 {
		log.Info("SendCoins", "节点数太少，停止空投", "nodes", noserviceCount+len(OnServiceIpMap))
		return
	}
	//留有缓冲检查的时间的时间
	time.Sleep(time.Second * 16)
	for addr := range toaddrs {
		var randCoin int64
		if ip, ok := OnseviceMap[addr]; ok {
			//提供服务的节点
			if pidarr, ok := OnServiceIpMap[ip]; ok {
				ReAvarage := float32(sendOnServiceCoins / float64(len(pidarr)))
				randCoin = int64(1000.0*ReAvarage) + int64(1000*sendNoSerViceCoins)
			}
		} else {
			randCoin = int64(sendNoSerViceCoins * 1000)
		}
		totalSend += float32(float32(randCoin) / 1000.0)
		log.Info("coins-------->", "addr", addr, "randCoin", randCoin, "txcount", txcount)

		for {
			if SendCoins(addr, 1e5*randCoin) != nil {
				//TODO发送失败的记录下来
				log.Error("sendError")
				time.Sleep(time.Second * 5)
				continue
			}
			break
		}

		txcount++
		time.Sleep(time.Millisecond * 500)

	}
	log.Info("BTY", "空投比特元结束", txcount, "totalsend", totalSend)
	return

}
