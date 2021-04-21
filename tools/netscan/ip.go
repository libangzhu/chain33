package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/p2p/dht"
	"github.com/33cn/chain33/tools/netscan/rpc"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var KvDb db.DB

//var path = `./datadir`
type schan chan string

var locaInfo *locationInfo

func initLocalInfo() {
	locaInfo = NewLocationInfo()
	KvDb = db.NewDB("netscan", "leveldb", ".", 128)

	//Reload LocalInfo

	var dhtgossipPath []string

	if *gossipPath != "" {
		log.Info("init","gossippath",*gossipPath)
		dhtgossipPath = append(dhtgossipPath, *gossipPath)
	}

	if *path != "" {
		dhtgossipPath = append(dhtgossipPath, *path)
	}


	for _, path := range dhtgossipPath {
		ReflushLocalInfo(path)
	}

	log.Info("init ok","size",locaInfo.Size())
	return
}

//reflush
func ReflushLocalInfo(path string) {
	var scan ScanFs
	var read ReadParase
	var foldersChan = make(schan, 128)

	now := time.Unix(time.Now().Unix(), 0)
	scan.ScanFolders(fmt.Sprintf("%v/DayRound_%v/", path, now.Format("2006-01-02")), foldersChan)
	for folder := range foldersChan {
		targetFolder := fmt.Sprintf("HourRound_%v", now.Hour())
		if folder != targetFolder {
			continue
		}
		rfb := read.ReadFile(fmt.Sprintf("%v/DayRound_%v/%v/countryinfos.txt", path, now.Format("2006-01-02"), folder))
		if len(rfb) == 0 {
			return
		}
		var cdata map[string][]string
		json.Unmarshal(rfb, &cdata)
		for k, v := range cdata {
			if _, ok := locaInfo.locationMap[k]; !ok {
				locaInfo.locationMap[k] = make(map[string]map[string]*cityInfo)
				locaInfo.locationMap[k]["region"] = make(map[string]*cityInfo)
				var city = new(cityInfo)
				city.pids=make(map[string]bool)
				for _,pid:=range v{
					city.pids[pid]=true
				}
				//city.pids = append(city.pids, v...)
				locaInfo.locationMap[k]["region"]["city"] = city
			}else{
				//add
				var ipinfo IP
				ipinfo.Country=k
				for _,pid:=range v{
					locaInfo.Add(&ipinfo,pid)
				}

			}
		}
		return
	}
}

func NewLocationInfo() *locationInfo {
	LocationInfo := new(locationInfo)
	LocationInfo.locationMap = make(map[string]map[string]map[string]*cityInfo)
	return LocationInfo
}

type cityInfo struct {
	//pids  []string
	pids map[string]bool
	point point
}
type locationInfo struct {
	mtx         sync.RWMutex
	Stat        bool
	locationMap map[string]map[string]map[string]*cityInfo
	data        interface{}
}

func (l *locationInfo) RemoveAll() {
	l.mtx.Lock()
	defer l.mtx.Unlock()

	for k, _ := range l.locationMap {
		delete(l.locationMap, k)
	}
	l.Stat = false
}

func (l *locationInfo) Size() int {
	var totalSize int
	for _, regionMap := range l.locationMap {

		var countryPidNum int
		for _, cityMap := range regionMap {
			for _, info := range cityMap {
				countryPidNum += len(info.pids)
			}
		}
		totalSize += countryPidNum
	}
	return totalSize
}

//Add
func (l *locationInfo) Add(ipinfo *IP, pid string) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	//pid check
	if _, err := hex.DecodeString(pid); err != nil {
		//try base58
		pubid, err := dht.PeerIDToPubkey(pid)
		if err != nil {
			return
		}
		pid = pubid
	}
	if _, ok := l.locationMap[ipinfo.Country]; !ok {
		l.locationMap[ipinfo.Country] = make(map[string]map[string]*cityInfo)
	}
	if ipinfo.Region == "" || ipinfo.Region == "XX" {
		ipinfo.Region = ipinfo.Country
	}
	if _, ok := l.locationMap[ipinfo.Country][ipinfo.Region]; !ok {
		l.locationMap[ipinfo.Country][ipinfo.Region] = make(map[string]*cityInfo)

	}
	if _,ok:=l.locationMap[ipinfo.Country]["region"];!ok{
		l.locationMap[ipinfo.Country]["region"] = make(map[string]*cityInfo)
	}

	if ipinfo.City == "" || ipinfo.City == "XX" {
		ipinfo.City = ipinfo.Region
	}
	if city, ok := l.locationMap[ipinfo.Country]["region"]["city"]; ok {
		city.pids[pid]=true
		//city.pids = append(city.pids, pid)
		l.locationMap[ipinfo.Country]["region"]["city"] = city
		return

	}
	var city = new(cityInfo)
	city.pids=make(map[string]bool)
	city.pids[pid]=true
	//city.pids = append(city.pids, pid)
	city.point = ipinfo.Point
	l.locationMap[ipinfo.Country]["region"]["city"] = city

}

type IPInfo struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

type IP struct {
	Country   string `json:"country"`
	CountryId string `json:"country_id"`
	Area      string `json:"area"`
	AreaId    string `json:"area_id"`
	Region    string `json:"region"`
	RegionId  string `json:"region_id"`
	City      string `json:"city"`
	CityId    string `json:"city_id"`
	Isp       string `json:"isp"`
	Point     point  `json:"point"`
}

type BaiDuIp struct {
	Address string  `json:"address"`
	Content content `json:"content"`
}

type content struct {
	AddressDetail interface{} `json:"address_detail"`
	Address       string      `json:"address"`
	Point         point       `json:"point"`
}

type point struct {
	X string `json:"x"`
	Y string `json:"y"`
}

func LocationAPI(ip string) *IPInfo {
	//url := "http://ip.taobao.com/service/getIpInfo.php?ip="
	//url += ip
	//log.Info("LocatoinAPi", "url", url)
	url := fmt.Sprintf("http://ip.taobao.com/outGetIpInfo?ip=%v&accessKey=alibaba-inc", ip)
	log.Info("LocatoinAPi", "url", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Error("http get", "err", err.Error())
		return nil
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("io read", "errr", err.Error())
		return nil
	}
	//log.Info("read", "out", string(out))
	var result IPInfo
	if err := json.Unmarshal(out, &result); err != nil {
		log.Error("json umarshal", "err", err.Error())
		return nil
	}

	return &result
}

func GetpeerLocaltionInfo() []*rpc.CountryInfo {
	var countryInfos []*rpc.CountryInfo
	locaInfo.mtx.Lock()
	defer locaInfo.mtx.Unlock()
	for country, regionData := range locaInfo.locationMap {
		var pidNum int
		for _, cityMap := range regionData {
			for _, info := range cityMap {
				pidNum += len(info.pids)
			}
		}
		countryinfo := "\n" + country + "node num:" + fmt.Sprintf("%v", pidNum)
		log.Info("GetPeerLocaltionInfo", "countryinfo", countryinfo)
		countryInfos = append(countryInfos, &rpc.CountryInfo{country, pidNum})
	}
	//read files

	return countryInfos
}
