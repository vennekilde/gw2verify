// DO NOT EDIT THIS FILE. This file will be overwritten when re-running go-raml.
package types

import (
	"gopkg.in/validator.v2"
)

type VerificationStatus struct {
	Account_id      string                       `json:"account_id,omitempty"`
	Attributes      []string                     `json:"attributes,omitempty"`
	Ban_reason      string                       `json:"ban_reason,omitempty"`
	Expires         int                          `json:"expires,omitempty"`
	Is_primary      bool                         `json:"is_primary,omitempty"`
	Primary_user_id string                       `json:"primary_user_id,omitempty"`
	Service_links   []ServiceLink                `json:"service_links,omitempty"`
	Status          EnumVerificationStatusStatus `json:"status" validate:"nonzero"`
}

func (s VerificationStatus) Validate() error {

	return validator.Validate(s)
}
