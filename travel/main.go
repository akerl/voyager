package travel

import (
	"fmt"
	"os"

	"github.com/akerl/voyager/cartogram"

	"github.com/akerl/speculate/creds"
	"github.com/akerl/speculate/executors"
)

type hop struct {
	Profile string
	Account string
	Role    string
	Mfa     bool
}

type voyage struct {
	pack    cartogram.Pack
	account cartogram.Account
	role    string
	hops    []hop
	creds   creds.Creds
}

// Itinerary describes a travel request
type Itinerary struct {
	Args        []string
	RoleName    string
	SessionName string
	Policy      string
	Lifetime    int64
	MfaCode     string
	MfaSerial   string
}

// TravelWithOptions loads creds from a full set of parameters
func Travel(i Itinerary) (creds.Creds, error) {
	var creds creds.Creds
	v := voyage{}

	if err := v.loadPack(); err != nil {
		return creds, err
	}
	if err := v.loadAccount(i.Args); err != nil {
		return creds, err
	}
	if err := v.loadRole(i.RoleName); err != nil {
		return creds, err
	}
	if err := v.loadHops(); err != nil {
		return creds, err
	}
	if err := v.loadCreds(i); err != nil {
		return creds, err
	}
	return v.creds, nil
}

func (v *voyage) loadPack() error {
	return v.pack.Load()
}

func (v *voyage) loadAccount(args []string) error {
	var err error
	v.account, err = v.pack.Find(args)
	return err
}

func (v *voyage) loadRole(roleName string) error {
	var err error
	v.role, err = v.account.PickRole(roleName)
	return err
}

func (v *voyage) loadHops() error {
	if err := parseHops(&v.hops, v.pack, v.account, v.role); err != nil {
		return err
	}
	for i, j := 0, len(v.hops)-1; i < j; i, j = i+1, j-1 {
		v.hops[i], v.hops[j] = v.hops[j], v.hops[i]
	}
	return nil
}

func (v *voyage) loadCreds(i Itinerary) error {
	var creds creds.Creds
	var err error

	profileHop, stack := v.hops[0], v.hops[1:]
	os.Setenv("AWS_PROFILE", profileHop.Profile)

	last := len(stack) - 1
	for index, thisHop := range stack {
		a := executors.Assumption{}
		if err := a.SetAccountID(thisHop.Account); err != nil {
			return err
		}
		if err := a.SetRoleName(thisHop.Role); err != nil {
			return err
		}
		if err := a.SetSessionName(i.SessionName); err != nil {
			return err
		}
		if err := a.SetLifetime(i.Lifetime); err != nil {
			return err
		}
		if index == last {
			if err := a.SetPolicy(i.Policy); err != nil {
				return err
			}
		}
		if thisHop.Mfa {
			if err := a.SetMfa(true); err != nil {
				return err
			}
			if err := a.SetMfaSerial(i.MfaSerial); err != nil {
				return err
			}
			if err := a.SetMfaCode(i.MfaCode); err != nil {
				return err
			}
		}
		creds, err = assumption.ExecuteWithCreds(creds)
		if err != nil {
			return err
		}
	}
	v.creds = creds
	return nil
}

func parseHops(stack *[]hop, cp cartogram.Pack, a cartogram.Account, r string) error {
	*stack = append(*stack, hop{Account: a.Account, Role: r, Mfa: a.Roles[r].Mfa})
	accountMatch := cartogram.AccountRegex.FindStringSubmatch(a.Source)
	if len(accountMatch) != 4 {
		*stack = append(*stack, hop{Profile: a.Source})
		return nil
	}
	sAccountID := accountMatch[1]
	sRole := accountMatch[3]
	found, sAccount := cp.Lookup(sAccountID)
	if !found {
		return fmt.Errorf("Failed to resolve hop for %s", sAccountID)
	}
	return parseHops(stack, cp, sAccount, sRole)
}
