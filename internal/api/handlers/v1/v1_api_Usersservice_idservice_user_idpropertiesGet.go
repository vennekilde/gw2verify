// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package v1

import (
	"encoding/json"
	"net/http"

	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
)

// Usersservice_idservice_user_idpropertiesGet is the handler for GET /v1/users/{service_id}/{service_user_id}/properties
// Get all user properties
func (api V1API) Usersservice_idservice_user_idpropertiesGet(w http.ResponseWriter, r *http.Request) {
	if apiservice.Permitted(w, r) == false {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var respBody []types.Property
	json.NewEncoder(w).Encode(&respBody)
}
