package utils

import (
	"fmt"
	"os"

	"github.com/akerl/voyager/cartogram"

	speculate "github.com/akerl/speculate/utils"
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
	creds   speculate.Creds
}

// SimpleGetNamedCreds provides you with creds based on an optional role, session name, and filter args
func SimpleGetNamedCreds(targetRole string, sessionName string, args []string) (speculate.Creds, error) {
	var creds speculate.Creds
	v := voyage{}

	if err := v.loadPack(); err != nil {
		return creds, err
	}
	if err := v.loadAccount(args); err != nil {
		return creds, err
	}
	if err := v.loadRole(targetRole); err != nil {
		return creds, err
	}
	if err := v.loadHops(); err != nil {
		return creds, err
	}
	if err := v.loadCreds(sessionName); err != nil {
		return creds, err
	}
	return v.creds, nil
}

// SimpleGetCreds provides you with creds based on an optional role and filter args
func SimpleGetCreds(targetRole string, args []string) (speculate.Creds, error) {
	return SimpleGetNamedCreds(targetRole, "", args)
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

func (v *voyage) loadCreds(sessionName string) error {
	var creds speculate.Creds
	var err error

	profileHop, stack := v.hops[0], v.hops[1:]
	os.Setenv("AWS_PROFILE", profileHop.Profile)

	for _, thisHop := range stack {
		assumption := speculate.Assumption{
			RoleName:    thisHop.Role,
			AccountID:   thisHop.Account,
			SessionName: sessionName,
		}
		assumption.Mfa.UseMfa = thisHop.Mfa
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
