package main

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/system/p2p/dht"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type schan chan string
type ReadParase struct {
}

func (r ReadParase) ReadFile(filepath string) []byte {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	rb, err := ioutil.ReadAll(f)
	if err != nil {
		return nil
	}
	return rb
}
func (r ReadParase) parseOnlineFileAddr(indata [][]byte) map[string]bool {
	var addrs =make(map[string]bool)
	for _, in := range indata {

		fileContent := string(in)
		st := strings.TrimSpace(string(fileContent))
		strs := strings.Split(st, "\n")
		for _, linestr := range strs {
			pidaddr := strings.Split(linestr, "@")
			if len(pidaddr) >= 2 {
				//通过pid生成地址
				//TODO 创建地址
				pub, err := common.FromHex(pidaddr[0])
				if err != nil {
					//may be dht
					//继续计算，有可能是DHT的PID格式
					pubstr, err := dht.PeerIDToPubkey(pidaddr[0])
					if err != nil {
						log.Info("PeerIDToPubkey", "err", err)
						continue
					}
					pub, err = common.FromHex(pubstr)
					if err != nil {
						log.Info("PeerIDToPubkey", "err", err)
						continue
					}

				}
				address := address.PubKeyToAddress(pub)

				//addrs = append(addrs, address.String())
				addrs[address.String()]=true
			}

		}
	}
	return addrs
}
func (r ReadParase) parseFileContentMap(indata [][]byte) (map[string]string, map[string][]string) {
	var addrData = make(map[string]string)
	var addrmap = make(map[string][]string)
	for _, in := range indata {
		fileContent := string(in)
		st := strings.TrimSpace(string(fileContent))
		strs := strings.Split(st, "\n")
		for _, linestr := range strs {
			pidaddr := strings.Split(linestr, "@")
			if len(pidaddr) >= 2 {
				//通过pid生成地址
				pub, err := common.FromHex(pidaddr[0])
				if err != nil {
					//继续计算，有可能是DHT的PID格式
					pubstr, err := dht.PeerIDToPubkey(pidaddr[0])
					if err != nil {
						log.Info("PeerIDToPubkey", "err", err)
						continue
					}
					pub, err = common.FromHex(pubstr)
					if err != nil {
						log.Info("PeerIDToPubkey", "err", err)
						continue
					}

				}
				address := address.PubKeyToAddress(pub)
				var ip string

				if strings.Contains(pidaddr[1], ":") {
					ipport := strings.Split(pidaddr[1], ":")
					ip = ipport[0]

				} else {
					ip = pidaddr[1]
				}
				addrData[address.String()] = ip

				if addrs, ok := addrmap[ip]; ok {
					addrs = append(addrs, address.String())
					addrmap[ip] = addrs
				} else {
					var addrs []string
					addrs = append(addrs, address.String())
					addrmap[ip] = addrs
				}

			}

		}
	}
	return addrData, addrmap
}

type ScanFs struct {
}

func (s ScanFs) ScanFolders(path string, folders schan) {
	//log.Info("scanfolders", "path", path)
	defer close(folders)
	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			folders <- f.Name()

		}
		return nil
	})

}

//解析文件数据
func (s ScanFs) ScanFolderFiles(path string, bucketOnline schan, bucketService schan) {
	//log.Info("ScanFolderFiles", "path", path)
	//defer close(bucketOnline)
	//defer close(bucketService)
	filepath.Walk(path,
		func(path string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}
			if f.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, "onlinepids.txt") {
				bucketOnline <- path
			} else if strings.HasSuffix(path, "onservice.txt") {
				bucketService <- path
			} else if strings.HasSuffix(path, "allNode.txt") {
				bucketOnline <- path
			}
			return nil
		})

}
