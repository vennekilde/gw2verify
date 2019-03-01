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

// Service_idservice_user_idapikeyPut is the handler for PUT /users/{service_id}/{service_user_id}/apikey
// Set a service user's API key
func (api UsersAPI) Service_idservice_user_idapikeyPut(w http.ResponseWriter, r *http.Request) {
	if apiservice.Permitted(w, r) == false {
		return
	}
	var reqBody types.APIKeyData

	// decode request
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(400)
		return
	}

	//URL Params
	params := mux.Vars(r)
	serviceID := params["service_id"]
	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		ThrowReqError(w, r, "service id is not an integer", http.StatusBadRequest)
		return
	}
	serviceUserID := params["service_user_id"]

	if reqBody.Apikey == "" {
		ThrowReqError(w, r, "apikey is missing", http.StatusBadRequest)
		return
	}

	//Check skip-requirements
	skipRequirements := false
	skipRequirementsList := r.URL.Query()["skip-requirements"]
	if len(skipRequirementsList) > 0 {
		skipRequirements, err = strconv.ParseBool(skipRequirementsList[0])
		if err != nil {
			skipRequirements = false
		}
	}

	gw2a := gw2api.NewGW2Api()
	err = verify.SetAPIKeyByUserService(gw2a, serviceIDInt, serviceUserID, reqBody.Primary, reqBody.Apikey, skipRequirements)
	if err != nil {
		ThrowReqError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
}
