// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package users

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
	"github.com/vennekilde/gw2verify/internal/config"
	"github.com/vennekilde/gw2verify/pkg/verify"
)

// Service_idservice_user_idverificationtemporaryPut is the handler for PUT /users/{service_id}/{service_user_id}/verification/temporary
// Grant a user temporary world relation
func (api UsersAPI) Service_idservice_user_idverificationtemporaryPut(w http.ResponseWriter, r *http.Request) {
	if apiservice.Permitted(w, r) == false {
		return
	}
	var reqBody types.TemporaryData
	worldPerspective := HARD_CODED_WORLD_PERSPECTIVE

	// decode request
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(400)
		return
	}

	var world int
	if reqBody.World > 0 {
		world = reqBody.World
	} else if len(reqBody.Access_type) > 0 {
		reqBody.Access_type = types.AccessType(strings.ToUpper(string(reqBody.Access_type)))
		if reqBody.Access_type == types.AccessTypeHOME_WORLD {
			// Grant Home World temporary access
			world = worldPerspective
		} else if reqBody.Access_type == types.AccessTypeLINKED_WORLD {
			// Grant Linked World temporary access
			worldLinks, err := verify.GetWorldLinks(worldPerspective)
			if err != nil {
				glog.Error(err)
				w.WriteHeader(400)
			}
			if len(worldLinks) > 0 {
				world = worldLinks[0]
			} else {
				// Not linked with another world, so cannot temporary grant linked world access
				// @TODO Consider just setting the user to home world temporary in this case
				ThrowReqError(w, r, "Currently not linked with any other servers", nil, http.StatusBadRequest)
				return
			}
		} else {
			ThrowReqError(w, r, "Invalid AccessType", nil, http.StatusBadRequest)
			return
		}
	} else {
		ThrowReqError(w, r, "Missing world or access_type", nil, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	//URL Params
	params := mux.Vars(r)
	serviceID := params["service_id"]
	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		ThrowReqError(w, r, "service id is not an integer", nil, http.StatusBadRequest)
		return
	}
	serviceUserID := params["service_user_id"]

	err, userErr := verify.GrantTemporaryWorldAssignment(serviceIDInt, serviceUserID, world)
	if err != nil {
		ThrowReqError(w, r, err.Error(), userErr, http.StatusInternalServerError)
		return
	}

	respBody := config.Config().TemporaryAccessExpirationTime
	json.NewEncoder(w).Encode(&respBody)
}
