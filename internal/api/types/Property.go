// DO NOT EDIT THIS FILE. This file will be overwritten when re-running go-raml.
package types

import (
	"gopkg.in/validator.v2"
)

type Property struct {
	Name  string `json:"name" validate:"nonzero"`
	Value string `json:"value" validate:"nonzero"`
}

func (s Property) Validate() error {

	return validator.Validate(s)
}
