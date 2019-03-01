// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package users

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

// UsersAPI is API implementation of /users root endpoint
type UsersAPI struct {
}

func ThrowReqError(w http.ResponseWriter, r *http.Request, errorMsg string, statusCode int) {
	jsonErr := make(map[string]string)
	jsonErr["error"] = errorMsg
	glog.Warningf("Request {URI: %s, RemoteAddr: %s} caused error msg: %s", r.RequestURI, r.RemoteAddr, errorMsg)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(&jsonErr)
}
