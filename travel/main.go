package travel

import (
	"fmt"
	"os"

	"github.com/akerl/speculate/creds"
	"github.com/akerl/speculate/executors"
	"github.com/akerl/voyager/cartogram"
	"github.com/akerl/voyager/prompt"
)

type hop struct {
	Profile string
	Account string
	Region  string
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
	Prompt      prompt.Func
}

// Travel loads creds from a full set of parameters
func Travel(i Itinerary) (creds.Creds, error) {
	var c creds.Creds
	v := voyage{}

	if i.Prompt == nil {
		i.Prompt = prompt.WithDefault
	}

	if err := v.loadPack(); err != nil {
		return c, err
	}
	if err := v.loadAccount(i.Args, i.Prompt); err != nil {
		return c, err
	}
	if err := v.loadRole(i.RoleName, i.Args, i.Prompt); err != nil {
		return c, err
	}
	if err := v.loadHops(); err != nil {
		return c, err
	}
	if err := v.loadCreds(i); err != nil {
		return c, err
	}
	return v.creds, nil
}

func (v *voyage) loadPack() error {
	v.pack = make(cartogram.Pack)
	return v.pack.Load()
}

func (v *voyage) loadAccount(args []string, pf prompt.Func) error {
	var err error
	v.account, err = v.pack.FindWithPrompt(args, pf)
	return err
}

func (v *voyage) loadRole(roleName string, args []string, pf prompt.Func) error {
	var err error
	if roleName == "" && len(args) == 1 {
		accountMatch := cartogram.AccountRegex.FindStringSubmatch(args[0])
		if len(accountMatch) > 2 {
			roleName = accountMatch[2]
		}
	}
	v.role, err = v.account.PickRoleWithPrompt(roleName, pf)
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
	var c creds.Creds
	var err error

	profileHop, stack := v.hops[0], v.hops[1:]
	for varName := range creds.Translations["envvar"] {
		err = os.Unsetenv(varName)
		if err != nil {
			return err
		}
	}
	err = os.Setenv("AWS_PROFILE", profileHop.Profile)
	if err != nil {
		return err
	}

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
		if i.Lifetime != 0 {
			if err := a.SetLifetime(i.Lifetime); err != nil {
				return err
			}
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
		c, err = a.ExecuteWithCreds(c)
		c.Region = thisHop.Region
		if err != nil {
			return err
		}
	}
	v.creds = c
	return nil
}

func parseHops(stack *[]hop, cp cartogram.Pack, a cartogram.Account, r string) error {
	*stack = append(
		*stack,
		hop{
			Account: a.Account,
			Region:  a.Region,
			Role:    r,
			Mfa:     a.Roles[r].Mfa,
		},
	)
	accountMatch := cartogram.AccountRegex.FindStringSubmatch(a.Source)
	if len(accountMatch) != 3 {
		*stack = append(*stack, hop{Profile: a.Source})
		return nil
	}
	sAccountID := accountMatch[1]
	sRole := accountMatch[2]
	found, sAccount := cp.Lookup(sAccountID)
	if !found {
		return fmt.Errorf("failed to resolve hop for %s", sAccountID)
	}
	return parseHops(stack, cp, sAccount, sRole)
}
