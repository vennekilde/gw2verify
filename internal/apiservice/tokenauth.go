package apiservice

import (
	"net/http"
	"strings"

	"github.com/vennekilde/gw2verify/internal/config"
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

	w.WriteHeader(http.StatusForbidden)
	return false
}
