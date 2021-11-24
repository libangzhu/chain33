package main

import (
	"flag"
	"fmt"

	logger "github.com/33cn/chain33/common/log/log15"
	"time"
)

var (
	log        = logger.New("module", "nodeaward")
	awardT     = flag.Int("t", 22, "nodeaward time")
	destpath   = flag.String("p", "", "scan file path")
	gossippath = flag.String("gp", "", "scan file path")
	chainver =flag.String("v","","chain33 version")
	commitId=flag.String("id","","commit id")
)

var waithChan chan bool

func main() {
	flag.Parse()
	fmt.Println("awardTime", *awardT)
ReStart:
	for {
		if time.Now().Hour() == *awardT { //每天22点时间，开始选择当天扫描数据进行空投
			var paths []string
			bestfloader := ScanBestFiles(*destpath)
			paths = append(paths, bestfloader)
			if *gossippath != "" {
				bestfloader = ScanBestFiles(*gossippath)
				paths = append(paths, bestfloader)
			}
			go normalAward(paths)
			break
		}

		time.Sleep(time.Second * 10)
	}
	time.Sleep(time.Hour * 3)
	goto ReStart
	<-waithChan
}
