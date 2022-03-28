package main

import (
	"fmt"
	"testing"
)

func Test_Parase(t *testing.T) {
	nt := &NetScan{}

	f, err := nt.ParaseRate("46.802 MB/s")
	if err != nil {
		t.Log(err)
		return
	}

	t.Log(fmt.Sprintf("%f", f))
}

