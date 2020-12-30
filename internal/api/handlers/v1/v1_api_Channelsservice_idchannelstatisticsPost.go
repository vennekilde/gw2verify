// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
	"github.com/vennekilde/gw2verify/pkg/history"
)

// Channelsservice_idchannelstatisticsPost is the handler for POST /v1/channels/{service_id}/{channel}/statistics
// Collect statistics based on the provided parameters and save them for historical purposes
func (api V1API) Channelsservice_idchannelstatisticsPost(w http.ResponseWriter, r *http.Request) {
	if apiservice.Permitted(w, r) == false {
		return
	}
	w.Header().Set("Content-Type", "application/json")

	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE

	//URL Params
	params := mux.Vars(r)
	serviceID := params["service_id"]
	ChannelID := params["channel"]
	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var reqBody types.ChannelMetadata
	// decode request
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(400)
		return
	}

	if err := history.CollectChannelStatistics(serviceIDInt, ChannelID, worldPerspective, reqBody); err != nil {
		ThrowReqError(w, r, err.Error(), err, http.StatusInternalServerError)
		return
	}
}
