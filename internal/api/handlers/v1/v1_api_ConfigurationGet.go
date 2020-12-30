// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package v1

import (
	"encoding/json"
	"net/http"

	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// ConfigurationGet is the handler for GET /v1/configuration
// Get a configuration containing relevant information for running a service bot
func (api V1API) ConfigurationGet(w http.ResponseWriter, r *http.Request) {
	if apiservice.Permitted(w, r) == false {
		return
	}
	links := verify.GetAllWorldLinks()
	values := make([]types.WorldLinks, 0, len(links))
	for _, v := range links {
		values = append(values, v)
	}

	w.Header().Set("Content-Type", "application/json")
	respBody := types.Configuration{
		Expiration_time:                  config.Config().ExpirationTime,
		Temporary_access_expiration_time: config.Config().TemporaryAccessExpirationTime,
		World_links:                      values,
	}
	json.NewEncoder(w).Encode(&respBody)
}
