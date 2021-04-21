package rpc

import (
	"compress/gzip"
	"fmt"
	log "github.com/33cn/chain33/common/log/log15"

	"github.com/33cn/chain33/queue"
	"github.com/rs/cors"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
)


type JSONRPCServer struct {
	jrpc       NetScan
	ListenAddr string
	client     queue.Client
}

// adapt HTTP connection to ReadWriteCloser
type HTTPConn struct {
	r   *http.Request
	in  io.Reader
	out io.Writer
}

func NewJRpcServer( port int,client queue.Client) *JSONRPCServer {
	jserver := new(JSONRPCServer)
	jserver.ListenAddr = fmt.Sprintf(":%d",port )
	jserver.client = client
	return jserver
}

func (c *HTTPConn) Read(p []byte) (n int, err error) { return c.in.Read(p) }

func (c *HTTPConn) Write(d []byte) (n int, err error) { //添加支持gzip 发送

	if strings.Contains(c.r.Header.Get("Accept-Encoding"), "gzip") {
		gw := gzip.NewWriter(c.out)
		defer gw.Close()
		return gw.Write(d)
	}
	return c.out.Write(d)
}

func (c *HTTPConn) Close() error { return nil }

func (j *JSONRPCServer) Listen() {
	listener, err := net.Listen("tcp", j.ListenAddr)
	if err != nil {
		log.Crit("listen:", "err", err)
		panic(err)
	}
	server := rpc.NewServer()
	j.jrpc.client = j.client
	server.Register(&j.jrpc)
	co := cors.New(cors.Options{})

	// Insert the middleware
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			serverCodec := jsonrpc.NewServerCodec(&HTTPConn{in: r.Body, out: w, r: r})
			w.Header().Set("Content-type", "application/json")
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				w.Header().Set("Content-Encoding", "gzip")
			}
			w.WriteHeader(200)
			err := server.ServeRequest(serverCodec)
			if err != nil {
				log.Debug("Error while serving JSON request: %v", err)
				return
			}
		}
	})

	handler = co.Handler(handler)
	http.Serve(listener, handler)
}
