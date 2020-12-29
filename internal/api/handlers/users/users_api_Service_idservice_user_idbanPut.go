// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package users

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// Service_idservice_user_idbanPut is the handler for PUT /users/{service_id}/{service_user_id}/ban
// Ban a user's gw2 account from being verified
func (api UsersAPI) Service_idservice_user_idbanPut(w http.ResponseWriter, r *http.Request) {
	if apiservice.Permitted(w, r) == false {
		return
	}
	//URL Params
	params := mux.Vars(r)
	serviceID := params["service_id"]
	serviceUserID := params["service_user_id"]
	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var reqBody types.BanData
	// decode request
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(400)
		return
	}

	duration := time.Duration(reqBody.Duration) * time.Millisecond
	err = verify.BanServiceUser(duration, reqBody.Reason, serviceIDInt, serviceUserID)
	if err != nil {
		ThrowReqError(w, r, err.Error(), err, http.StatusInternalServerError)
	} else {
		w.WriteHeader(200)
	}
}
