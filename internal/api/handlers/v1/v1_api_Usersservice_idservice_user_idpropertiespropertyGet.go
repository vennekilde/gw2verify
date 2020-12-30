// THIS FILE IS SAFE TO EDIT. It will not be overwritten when rerunning go-raml.
package v1

import (
	"encoding/json"
	"net/http"

	"github.com/vennekilde/gw2verify/internal/api/types"
	"github.com/vennekilde/gw2verify/internal/apiservice"
)

// Usersservice_idservice_user_idpropertiespropertyGet is the handler for GET /v1/users/{service_id}/{service_user_id}/properties/{property}
// Get a user property
func (api V1API) Usersservice_idservice_user_idpropertiespropertyGet(w http.ResponseWriter, r *http.Request) { // name := req.FormValue("name")
	if apiservice.Permitted(w, r) == false {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	var respBody types.Property
	json.NewEncoder(w).Encode(&respBody)
}
