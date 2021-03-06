// DO NOT EDIT THIS FILE. This file will be overwritten when re-running go-raml.
package types

import (
	"gopkg.in/validator.v2"
)

type Configuration struct {
	Expiration_time                  int          `json:"expiration_time" validate:"nonzero"`
	Temporary_access_expiration_time int          `json:"temporary_access_expiration_time" validate:"nonzero"`
	World_links                      []WorldLinks `json:"world_links" validate:"nonzero"`
}

func (s Configuration) Validate() error {

	return validator.Validate(s)
}
