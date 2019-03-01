// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package configuration

import (
	"encoding/json"
	"net/http"

	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// Get is the handler for GET /configuration
// Get a configuration containing relevant information for running a service bot
func (api ConfigurationAPI) Get(w http.ResponseWriter, r *http.Request) {
	if apiservice.Permitted(w, r) == false {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	respBody := types.Configuration{
		Expiration_time:                  config.Config().ExpirationTime,
		Temporary_access_expiration_time: config.Config().TemporaryAccessExpirationTime,
		Home_world:                       config.Config().HomeWorld,
		Link_worlds:                      verify.Config.LinkedWorlds,
	}
	json.NewEncoder(w).Encode(&respBody)
}
