// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package users

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// Service_idservice_user_idapikeynameGet is the handler for GET /users/{service_id}/{service_user_id}/apikey/name
// Get a service user's API key name
func (api UsersAPI) Service_idservice_user_idapikeynameGet(w http.ResponseWriter, r *http.Request) {
	if apiservice.Permitted(w, r) == false {
		return
	}

	//URL Params
	params := mux.Vars(r)
	serviceID := params["service_id"]
	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		ThrowReqError(w, r, "service id is not an integer", nil, http.StatusBadRequest)
		return
	}
	serviceUserID := params["service_user_id"]

	apikeyName := verify.GetAPIKeyName(serviceIDInt, serviceUserID)
	RespWithSuccess(w, r, types.APIKeyName{Name: apikeyName})
}
