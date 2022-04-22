// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
	"github.com/vennekilde/gw2verify/pkg/sync"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// Usersservice_idservice_user_idverificationrefreshPost is the handler for POST /v1/users/{service_id}/{service_user_id}/verification/refresh
// Forces a refresh of the API data and returns the new verification status after the API data has been refreshed. Note this can take a few seconds
func (api V1API) Usersservice_idservice_user_idverificationrefreshPost(w http.ResponseWriter, r *http.Request) {
	if apiservice.Permitted(w, r) == false {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE

	//URL Params
	params := mux.Vars(r)
	serviceID := params["service_id"]
	serviceUserID := params["service_user_id"]
	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	gw2a := gw2api.NewGW2Api()
	err, userErr := sync.SynchronizeLinkedUser(gw2a, serviceIDInt, serviceUserID)
	if err != nil {
		ThrowReqError(w, r, err.Error(), userErr, http.StatusInternalServerError)
		return
	}
	status, _, err := verify.Status(worldPerspective, serviceIDInt, serviceUserID)
	if err != nil {
		ThrowReqError(w, r, err.Error(), err, http.StatusInternalServerError)
		return
	}
	var respBody types.VerificationStatus
	respBody.Status = types.EnumVerificationStatusStatus(status.Status.Name())
	respBody.Account_id = status.AccountData.ID
	respBody.Expires = status.Expires
	if status.Status == verify.ACCESS_DENIED_BANNED {
		respBody.Ban_reason = status.Description
	}

	serviceListener := verify.ServicePollListeners[serviceIDInt]
	if serviceListener.Listener != nil {
		serviceListener.Listener <- respBody
	}

	json.NewEncoder(w).Encode(&respBody)
}
