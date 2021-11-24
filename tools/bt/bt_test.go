package bt

import (
	"github.com/zeebo/bencode"
	"reflect"
	"testing"
	"unsafe"
)
//https://github.com/zeebo/bencode


func Test_Bencode(t *testing.T){
	encodeStr,err:=bencode.EncodeString("121212")
	if err!=nil{
		t.Log("err:",err.Error())
		return
	}

	t.Log("encodeStr:",encodeStr)
	eb,err:= bencode.EncodeBytes([]byte("hello1234"))
	if err!=nil{
		t.Log("err:",err.Error())
		return
	}

	t.Log("en:",string(eb))

	var a int32=16
	var buf[4]byte
	p:=(*int32)(unsafe.Pointer(&buf))
	*p=a
	eb,err= bencode.EncodeBytes(buf[:])
	if err!=nil{
		t.Log("err:",err.Error())
		return
	}

	t.Log("en:",eb)
	t.Log("type:",reflect.TypeOf(p))
	var v interface{}
	bencode.DecodeBytes(eb,&v)
	t.Log("v:",v)

}


func Test_EncodeString(t *testing.T) {
	var torrent interface{}
	data, err := bencode.EncodeString(torrent)
	if err != nil {
		panic(err)
	}
	t.Log("data:",data)
}