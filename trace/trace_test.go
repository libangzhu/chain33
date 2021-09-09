package trace

import "testing"
import (
	"net/http"
)

func TestNew(t *testing.T) {

		service:=	New(nil)
		err:= http.ListenAndServe("localhost:33010",service)
		if err!=nil{
			t.Log(err.Error())
		}

}