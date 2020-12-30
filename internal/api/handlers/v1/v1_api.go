// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package v1

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

//@TODO fix hardcoded later
var HARD_CODED_WORLD_PERSPECTIVE = 2007

// V1API is API implementation of /v1 root endpoint
type V1API struct {
}

func ThrowReqError(w http.ResponseWriter, r *http.Request, errorMsg string, userErr error, statusCode int) {
	jsonErr := make(map[string]string)
	jsonErr["error"] = errorMsg
	if userErr != nil {
		jsonErr["safe-display-error"] = userErr.Error()
	}
	glog.Warningf("Request {URI: %s, RemoteAddr: %s} caused error msg: %s", r.RequestURI, r.RemoteAddr, errorMsg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(&jsonErr)
}

func RespWithSuccess(w http.ResponseWriter, r *http.Request, respBody interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&respBody)
}
