// DO NOT EDIT THIS FILE. This file will be overwritten when re-running go-raml.
package types

import (
	"gopkg.in/validator.v2"
)

type ServiceLink struct {
	Display_name    string `json:"display_name,omitempty"`
	Service_id      int    `json:"service_id" validate:"nonzero"`
	Service_user_id string `json:"service_user_id" validate:"nonzero"`
}

func (s ServiceLink) Validate() error {

	return validator.Validate(s)
}
