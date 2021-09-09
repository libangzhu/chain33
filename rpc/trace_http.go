package rpc

import (
	"github.com/33cn/chain33/trace"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)




func RunTraceServer(){
	debugAPIListener, err := net.Listen("tcp", ":12345")
	if err!=nil{
		panic(err)
	}
	debugAPIServer := &http.Server{
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           trace.New(),
		ErrorLog:          log.New(os.Stderr,"",0),
	}

	debugAPIServer.Serve(debugAPIListener)


}