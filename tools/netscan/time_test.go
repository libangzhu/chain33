package main

import (
	"fmt"
	"testing"
	"time"
)

func Test_paraseT(t *testing.T){

	now := fmt.Sprintf("%v-%v", time.Now().Format("2006-01-02"), time.Now().Hour())

	t.Log("now:",now,"key:",NetRatePrefix+now)
}
