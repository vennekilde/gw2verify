package sync

import "github.com/vennekilde/gw2verify/v2/internal/api"

// Contains checks if a slice contains the given item
func Contains(slice []string, item string) bool {
	for _, itemInSlice := range slice {
		if itemInSlice == item {
			return true
		}
	}
	return false
}

// IsFreeToPlay returns true if the account is a free to play account
func IsFreeToPlay(access []string) bool {
	return Contains(access, api.PlayForFree) && !Contains(access, api.GuildWars2)
}
