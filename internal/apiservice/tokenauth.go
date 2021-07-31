package apiservice

import (
	"net/http"
	"strings"

	"github.com/vennekilde/gw2verify/internal/config"
	"go.uber.org/zap"
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

	zap.L().Warn("unable to verify token from request",
		zap.String("request uri", r.RequestURI),
		zap.String("remote addr", r.RemoteAddr),
		zap.String("token", token))
	w.WriteHeader(http.StatusForbidden)
	return false
}
