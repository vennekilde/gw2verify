// DO NOT EDIT THIS FILE. This file will be overwritten when re-running go-raml.
package types

import (
	"gopkg.in/validator.v2"
)

type ChannelUserMetadata struct {
	Deafened  bool   `json:"deafened"`
	Id        string `json:"id" validate:"nonzero"`
	Muted     bool   `json:"muted"`
	Name      string `json:"name" validate:"nonzero"`
	Streaming bool   `json:"streaming"`
}

func (s ChannelUserMetadata) Validate() error {

	return validator.Validate(s)
}