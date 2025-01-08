package api

import (
	"strings"

	"github.com/MrGunflame/gw2api"
)

const (
	ACCESS_DENIED_ACCOUNT_NOT_LINKED      = ACCESSDENIEDACCOUNTNOTLINKED
	ACCESS_DENIED_BANNED                  = ACCESSDENIEDBANNED
	ACCESS_DENIED_EXPIRED                 = ACCESSDENIEDEXPIRED
	ACCESS_DENIED_INVALID_WORLD           = ACCESSDENIEDINVALIDWORLD
	ACCESS_DENIED_REQUIREMENT_NOT_MET     = ACCESSDENIEDREQUIREMENTNOTMET
	ACCESS_DENIED_UNKNOWN                 = ACCESSDENIEDUNKNOWN
	ACCESS_GRANTED_HOME_WORLD             = ACCESSGRANTEDHOMEWORLD
	ACCESS_GRANTED_HOME_WORLD_TEMPORARY   = ACCESSGRANTEDHOMEWORLDTEMPORARY
	ACCESS_GRANTED_LINKED_WORLD           = ACCESSGRANTEDLINKEDWORLD
	ACCESS_GRANTED_LINKED_WORLD_TEMPORARY = ACCESSGRANTEDLINKEDWORLDTEMPORARY
)
const (
	HOME_WORLD   = HOMEWORLD
	LINKED_WORLD = LINKEDWORLD
)

func (s *VerificationStatus) WithStatus(status Status) *VerificationStatus {
	s.Status = status
	return s
}

func (s *VerificationStatus) WithStatusIfHigher(status Status) *VerificationStatus {
	if status.Priority() > s.Status.Priority() {
		s.Status = status
	}
	return s
}

func (s Status) ID() int {
	switch s {
	case ACCESS_DENIED_UNKNOWN:
		return 0
	case ACCESS_GRANTED_HOME_WORLD:
		return 1
	case ACCESS_GRANTED_LINKED_WORLD:
		return 2
	case ACCESS_GRANTED_HOME_WORLD_TEMPORARY:
		return 3
	case ACCESS_GRANTED_LINKED_WORLD_TEMPORARY:
		return 4
	case ACCESS_DENIED_ACCOUNT_NOT_LINKED:
		return 5
	case ACCESS_DENIED_EXPIRED:
		return 6
	case ACCESS_DENIED_INVALID_WORLD:
		return 7
	case ACCESS_DENIED_BANNED:
		return 8
	case ACCESS_DENIED_REQUIREMENT_NOT_MET:
		return 9
	default:
		return -1
	}
}

func (s Status) Priority() int {
	switch s {
	case ACCESSDENIEDBANNED:
		return 100
	case ACCESSGRANTEDHOMEWORLD:
		return 90
	case ACCESSGRANTEDLINKEDWORLD:
		return 80
	case ACCESSGRANTEDHOMEWORLDTEMPORARY:
		return 70
	case ACCESSGRANTEDLINKEDWORLDTEMPORARY:
		return 60
	case ACCESSDENIEDINVALIDWORLD:
		return 50
	case ACCESSDENIEDEXPIRED:
		return 40
	case ACCESSDENIEDACCOUNTNOTLINKED:
		return 30
	case ACCESSDENIEDREQUIREMENTNOTMET:
		return 20
	case ACCESSDENIEDUNKNOWN:
		return 10
	default:
		return 0
	}
}

func (s Status) AccessGranted() bool {
	return strings.HasPrefix(string(s), "ACCESS_GRANTED")
}

func (s Status) AccessDenied() bool {
	return !s.AccessGranted()
}

func (acc *Account) FromGW2API(gw2Acc gw2api.Account) {
	acc.Access = &gw2Acc.Access
	acc.Age = int(gw2Acc.Age)
	acc.Commander = gw2Acc.Commander
	acc.Created = gw2Acc.Created
	acc.DailyAp = &gw2Acc.DailyAP
	acc.FractalLevel = &gw2Acc.FractalLevel
	if len(gw2Acc.GuildLeader) > 0 {
		acc.GuildLeader = &gw2Acc.GuildLeader
	}
	if len(gw2Acc.Guilds) > 0 {
		acc.Guilds = &gw2Acc.Guilds
	}
	acc.ID = gw2Acc.ID
	acc.MonthlyAp = &gw2Acc.MonthlyAP
	acc.Name = gw2Acc.Name
	acc.World = gw2Acc.World
	acc.WvWRank = gw2Acc.WvWRank
	acc.WvWTeamID = gw2Acc.WvW.TeamID
	acc.LastModified = &gw2Acc.LastModified
}
