// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package users

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vennekilde/gw2verify/pkg/verify"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
)

// Service_idservice_user_idverificationstatusGet is the handler for GET /users/{service_id}/{service_user_id}/verification/status
// Get a users verification status
func (api UsersAPI) Service_idservice_user_idverificationstatusGet(w http.ResponseWriter, r *http.Request) { // display_name := req.FormValue("display_name")
	if apiservice.Permitted(w, r) == false {
		return
	}

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
	status := verify.Status(serviceIDInt, serviceUserID)
	var respBody types.VerificationStatus
	respBody.Status = types.EnumVerificationStatusStatus(status.Status.Name())
	respBody.Account_id = status.AccountID
	respBody.Expires = status.Expires

	json.NewEncoder(w).Encode(&respBody)
}
