package verify

import "gitlab.com/MrGunflame/gw2api"

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
func IsFreeToPlay(acc gw2api.Account) bool {
	return Contains(acc.Access, PlayForFree) && !Contains(acc.Access, GuildWars2)
}
