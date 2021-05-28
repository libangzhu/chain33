package main

import (
	"fmt"
	"strings"
	"time"
)

func ScanBestFiles(path string) string {
	//扫描目录下的文件，找出最好的结果
	var foldersChan = make(schan, 25)
	now := time.Unix(time.Now().Unix(), 0) //获取当前时间
	currentPath := fmt.Sprintf("%v/DayRound_%v/", path, now.Format("2006-01-02"))
	log.Info("initpath", "path", currentPath)
	if path != "" {
		//currentPath=*destpath
		if strings.Contains(path, "DayRound") {
			currentPath = path
		} else {
			currentPath = fmt.Sprintf("%v/DayRound_%v/", path, now.Format("2006-01-02"))
		}

	}

	var scanfs ScanFs
	scanfs.ScanFolders(currentPath, foldersChan)
	var maxNum int
	var bestfolder string
	var read ReadParase
	for folder := range foldersChan {
		//log.Info("ScanBestFiles", "scanFolders", folder)
		var onlineChans = make(schan, 128)
		var onserviceChans = make(schan, 128)

		scanfs.ScanFolderFiles(currentPath+folder, onlineChans, onserviceChans)
		close(onlineChans)
		close(onserviceChans)
		var OnseviceMap map[string]string
		for filename := range onserviceChans {
			rbs := read.ReadFile(filename)
			OnseviceMap, _ = read.parseOnServFileContentMap([][]byte{rbs})
			if maxNum < len(OnseviceMap) {
				maxNum = len(OnseviceMap)
				//bestFile = filename
				bestfolder = folder
			}

		}

	}

	log.Info("best file", "path", currentPath+bestfolder)
	return currentPath + bestfolder
}
