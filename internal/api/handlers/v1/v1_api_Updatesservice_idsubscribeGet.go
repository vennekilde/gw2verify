// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package v1

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
)

type VerifictionStatusListener struct {
	WorldPerspective int
	ServiceID        int
	Listener         chan types.VerificationStatus
}

var ServicePollListeners map[int]VerifictionStatusListener = make(map[int]VerifictionStatusListener)

// Updatesservice_idsubscribeGet is the handler for GET /v1/updates/{service_id}/subscribe
// Long polling rest endpoint for receiving verification updates
func (api V1API) Updatesservice_idsubscribeGet(w http.ResponseWriter, r *http.Request) {
	if apiservice.Permitted(w, r) == false {
		return
	}
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

	statusListener := ServicePollListeners[serviceIDInt]
	if statusListener.Listener == nil {
		statusListener = VerifictionStatusListener{
			ServiceID:        serviceIDInt,
			WorldPerspective: HARD_CODED_WORLD_PERSPECTIVE,
			Listener:         make(chan types.VerificationStatus),
		}
		ServicePollListeners[serviceIDInt] = statusListener
	}

	select {
	case event := <-statusListener.Listener:
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(&event)
	case <-ticker.C:
		w.WriteHeader(408)
	}
	statusListener.Listener = nil
}
