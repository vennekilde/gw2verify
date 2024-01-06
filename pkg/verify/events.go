package verify

import (
	"github.com/vennekilde/gw2verify/v2/internal/api"
	"go.uber.org/zap"
)

type VerificationStatusListener struct {
	ID               string
	WorldPerspective int
	PlatformID       *int
	ch               chan *api.VerificationStatus
}

type UserEventListener struct {
	ID         string
	PlatformID *int
	ch         chan *api.User
}

var PlatformPollListeners map[int]VerificationStatusListener = make(map[int]VerificationStatusListener)

type EventEmitter struct {
	verification *Verification

	userListeners   map[string]*UserEventListener
	statusListeners map[string]*VerificationStatusListener
}

func NewEventEmitter(verification *Verification) *EventEmitter {
	return &EventEmitter{
		verification:    verification,
		userListeners:   make(map[string]*UserEventListener),
		statusListeners: make(map[string]*VerificationStatusListener),
	}
}

func (em *EventEmitter) GetUserListener(id string, platformID *int) chan *api.User {
	listener, ok := em.userListeners[id]
	if !ok {
		listener = &UserEventListener{
			ID:         id,
			PlatformID: platformID,
			ch:         make(chan *api.User, 10),
		}
		em.userListeners[id] = listener
	}

	return listener.ch
}

func (em *EventEmitter) GetStatusListener(id string, platformID *int, worldPerspective int) chan *api.VerificationStatus {
	listener, ok := em.statusListeners[id]
	if !ok {
		listener = &VerificationStatusListener{
			ID:               id,
			PlatformID:       platformID,
			WorldPerspective: worldPerspective,
			ch:               make(chan *api.VerificationStatus, 10),
		}
		em.statusListeners[id] = listener
	}

	return listener.ch
}

func (em *EventEmitter) Emit(user *api.User) {
	// Emit user updates
	for _, listener := range em.userListeners {
		if listener.PlatformID != nil {
			// Check if user is on the platform
			for _, link := range user.PlatformLinks {
				if link.PlatformID == *listener.PlatformID {
					goto exitLinkCheck
				}
			}
			// If we got here, it means the user is not on the platform
			continue
		}
	exitLinkCheck:

		select {
		case listener.ch <- user:
		default:
			zap.L().Error("unable to send verification update to listener", zap.Any("listener", listener))
		}
	}

	// Emit verification updates
	var platformLink *api.PlatformLink
	for _, listener := range em.statusListeners {
		if listener.PlatformID != nil {
			// Check if user is on the platform
			for _, link := range user.PlatformLinks {
				if link.PlatformID == *listener.PlatformID {
					platformLink = &link
					goto exitLinkCheck2
				}
			}
			// If we got here, it means the user is not on the platform
			continue
		}
	exitLinkCheck2:

		status := api.VerificationStatus{
			Status:       em.verification.Status(listener.WorldPerspective, user),
			Ban:          GetActiveBan(user.Bans),
			PlatformLink: platformLink,
		}

		select {
		case listener.ch <- &status:
		default:
			zap.L().Error("unable to send verification update to listener", zap.Any("listener", listener))
		}
	}
}

func (em *EventEmitter) Process(oldUser *api.User, newUser *api.User) {
	shouldEmit := false

	if oldUser == nil && newUser != nil {
		shouldEmit = true
	} else if len(oldUser.Bans) != len(newUser.Bans) {
		shouldEmit = true
	} else if em.ShouldEmitAccounts(oldUser.Accounts, newUser.Accounts) {
		shouldEmit = true
	} else if em.ShouldEmitEphemeralAssociations(oldUser.EphemeralAssociations, newUser.EphemeralAssociations) {
		shouldEmit = true
	}

	if shouldEmit {
		em.Emit(newUser)
	}
}

func (em *EventEmitter) ShouldEmitAccounts(oldAccounts []api.Account, newAccounts []api.Account) bool {
	if len(oldAccounts) != len(newAccounts) {
		return true
	}

outerLoop:
	for i := range newAccounts {
		newAccount := &newAccounts[i]
		for k := range oldAccounts {
			oldAccount := &oldAccounts[k]
			if oldAccount.Id == newAccount.Id {
				if em.ShouldEmitAccount(oldAccount, newAccount) {
					return true
				}
				continue outerLoop
			}
		}
		// If this is reached, we didn't find a matching account
		return true
	}

	return false
}

func (em *EventEmitter) ShouldEmitAccount(oldAccount *api.Account, newAccount *api.Account) bool {
	if oldAccount.World != newAccount.World {
		return true
	} else if oldAccount.Name != newAccount.Name {
		return true
	} else if newAccount.Expired != nil && *newAccount.Expired {
		return true
	}
	return false
}

func (em *EventEmitter) ShouldEmitEphemeralAssociations(olds []api.EphemeralAssociation, news []api.EphemeralAssociation) bool {
	if len(olds) != len(news) {
		return true
	}

outerLoop:
	for i := range news {
		new := &news[i]
		for k := range olds {
			old := &olds[k]
			if old.Until == new.Until && old.World == new.World {
				continue outerLoop
			}
		}
		// If this is reached, we didn't find a matching EphemeralAssociation
		return true
	}

	return false
}
