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

type Voyage struct {
	pack cartogram.Pack
	account cartogram.Account
	role string
	hops []hop
	creds speculate.Creds
}

// SimpleGetCreds provides you with creds based on an optional role and filter args
func SimpleGetCreds(targetRole string, args []string) (speculate.Creds, error) {
	var creds speculate.Creds
	v := Voyage{}

	if err := v.LoadPack(); err != nil {
		return creds, err
	}

    if err := v.LoadAccount(args); err != nil {
        return creds, err
    }

	if err := v.LoadRole(targetRole); err != nil {
        return creds, err
    }

    if err := v.LoadHops(); err != nil {
        return creds, err
    }
}

func (v *Voyage) LoadPack() error {
    return v.pack.Load()
}

func (v *Voyage) LoadAccount(args []string) error {
	v.account, err := v.pack.Find(args)
	return err
}

func (v *Voyage) LoadRole(roleName string) error {
	v.role, err = targetAccount.PickRole(targetRole)
	return err
}

func (v *Voyage) LoadHops() error {
    if err := parseHops(&v.hops, cp, targetAccount, targetRole); err != nil {
        return err
    }
    for i, j := 0, len(v.hops)-1; i < j; i, j = i+1, j-1 {
        v.hops[i], v.hops[j] = v.hops[j], v.hops[i]
    }
	return nil
}

func (v *Voyage) LoadCreds() (speculate.Creds)

// Travel accepts a role and args and turns them into creds
func Travel(targetRole string, args []string) (speculate.Creds, error) {
	return NamedTravel(targetRole, args, "")
}

// NamedTravel accepts a role, args, and session name and turns them into creds
func NamedTravel(targetRole string, args []string, sessionName string) (speculate.Creds, error) {
	cp := cartogram.Pack{}
	if err := cp.Load(); err != nil {
		return creds, err
	}

	targetAccount, err := cp.Find(args)
	if err != nil {
		return creds, err
	}

	targetRole, err = targetAccount.PickRole(targetRole)
	if err != nil {
		return creds, err
	}

	stack := []hop{}
	if err := parseHops(&stack, cp, targetAccount, targetRole); err != nil {
		return creds, err
	}
	for i, j := 0, len(stack)-1; i < j; i, j = i+1, j-1 {
		stack[i], stack[j] = stack[j], stack[i]
	}

	profileHop, stack := stack[0], stack[1:]
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
			return creds, err
		}
	}

	return creds, nil
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
