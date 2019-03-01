// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package updates

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/vennekilde/gw2verify/internal/api/types"
)

var ServicePollListeners map[int]chan types.VerificationStatus = make(map[int]chan types.VerificationStatus)

// Service_idsubscribeGet is the handler for GET /updates/{service_id}/subscribe
// Long polling rest endpoint for receiving verification updates
func (api UpdatesAPI) Service_idsubscribeGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	serviceID := params["service_id"]
	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ticker := time.NewTicker(120 * time.Second)
	defer func() { ticker.Stop() }()

	if ServicePollListeners[serviceIDInt] == nil {
		ServicePollListeners[serviceIDInt] = make(chan types.VerificationStatus, 100)
	}

	for {
		select {
		case event := <-ServicePollListeners[serviceIDInt]:
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(&event)
			return
		case <-ticker.C:
			w.WriteHeader(408)
			return
		}
	}

}
