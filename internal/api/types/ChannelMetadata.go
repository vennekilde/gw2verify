// DO NOT EDIT THIS FILE. This file will be overwritten when re-running go-raml.
package types

import (
	"gopkg.in/validator.v2"
)

type ChannelMetadata struct {
	Name  string                `json:"name" validate:"nonzero"`
	Users []ChannelUserMetadata `json:"users" validate:"nonzero"`
}

func (s ChannelMetadata) Validate() error {

	return validator.Validate(s)
}