package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func Createfile(filename string) *os.File {

	folderName := fmt.Sprintf("datadir/DayRound_%v/HourRound_%v", time.Now().Format("2006-01-02"), time.Now().Hour())
	if err := os.MkdirAll(folderName, 0755); err != nil {
		log.Error("create folder", "err", err.Error())
		return nil
	}
	if err := os.MkdirAll(folderName, 0755); err != nil {
		log.Error("create folder", "err", err.Error())
		return nil
	}
	pf, err := os.Create(fmt.Sprintf("%v/%v.txt", folderName, filename))
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return pf
}

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

type ScanFs struct {
}

func (s ScanFs) ScanFolders(path string, folders schan) {
	log.Info("scanfolders", "path", path)
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
