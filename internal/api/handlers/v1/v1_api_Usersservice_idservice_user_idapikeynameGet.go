// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package v1

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// Usersservice_idservice_user_idapikeynameGet is the handler for GET /v1/users/{service_id}/{service_user_id}/apikey/name
// Get a service user's apikey name they are required to use if apikey name restriction is enforced
func (api V1API) Usersservice_idservice_user_idapikeynameGet(w http.ResponseWriter, r *http.Request) {
	if apiservice.Permitted(w, r) == false {
		return
	}
	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE

	//URL Params
	params := mux.Vars(r)
	serviceID := params["service_id"]
	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		ThrowReqError(w, r, "service id is not an integer", nil, http.StatusBadRequest)
		return
	}
	serviceUserID := params["service_user_id"]

	apikeyName := verify.GetAPIKeyName(worldPerspective, serviceIDInt, serviceUserID)
	RespWithSuccess(w, r, types.APIKeyName{Name: apikeyName})
}
