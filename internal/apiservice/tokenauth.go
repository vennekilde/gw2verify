package apiservice

import (
	"net/http"
	"strings"

	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/golang/glog"
)

type AuthToken struct {
	Token string
}

func Permitted(w http.ResponseWriter, r *http.Request) bool {
	token := strings.ToUpper(r.Header.Get("X-Access-Token"))
	if token != "" {
		//for _, accessToken := range config.Config().RESTAuthTokens {
		accessToken := config.Config().RESTAuthToken
		if token == strings.ToUpper(accessToken) {
			return true
		}
		//}
	}

	glog.Warningf("Could not verify token on request {URI: %s, RemoteAddr: %s} token: %s", r.RequestURI, r.RemoteAddr, token) 
	w.WriteHeader(http.StatusForbidden)
	return false
}
