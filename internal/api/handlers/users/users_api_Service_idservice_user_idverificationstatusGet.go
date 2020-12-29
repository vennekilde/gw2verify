// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package users

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vennekilde/gw2apidb/pkg/gw2api"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

type VerificationStatusExtended struct {
	types.VerificationStatus
	AccountData gw2api.Account
}

// Service_idservice_user_idverificationstatusGet is the handler for GET /users/{service_id}/{service_user_id}/verification/status
// Get a users verification status
func (api UsersAPI) Service_idservice_user_idverificationstatusGet(w http.ResponseWriter, r *http.Request) { // display_name := req.FormValue("display_name")
	if apiservice.Permitted(w, r) == false {
		return
	}
	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE

	w.Header().Set("Content-Type", "application/json")
	//URL Params
	params := mux.Vars(r)
	serviceID := params["service_id"]
	serviceUserID := params["service_user_id"]
	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	status, link, err := verify.Status(worldPerspective, serviceIDInt, serviceUserID)
	if err != nil {
		ThrowReqError(w, r, err.Error(), err, http.StatusInternalServerError)
		return
	}
	linkREST := types.ServiceLink{
		Display_name:    link.ServiceUserDisplayName,
		Service_id:      link.ServiceID,
		Service_user_id: link.ServiceUserID,
	}
	respBody := VerificationStatusExtended{
		VerificationStatus: types.VerificationStatus{
			Status:        types.EnumVerificationStatusStatus(status.Status.Name()),
			Account_id:    status.AccountData.ID,
			Expires:       status.Expires,
			Service_links: []types.ServiceLink{linkREST},
		},
		AccountData: status.AccountData,
	}

	if status.Status == verify.ACCESS_DENIED_BANNED {
		respBody.Ban_reason = status.Description
	}
	json.NewEncoder(w).Encode(&respBody)
}
