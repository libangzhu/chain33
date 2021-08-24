package trace

import (
	"github.com/33cn/chain33/common/log/log15"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"sync"
)
var log = log15.New("module","trace")

type Service struct {
	metricsRegistry    *prometheus.Registry
	// handler is changed in the Configure method
	handler   http.Handler
	handlerMu sync.RWMutex

}

func New()*Service{
	s := new(Service)


}


