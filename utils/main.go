package utils

import (
	"fmt"
	"os"
	"regexp"
	"sort"

	"github.com/akerl/voyager/cartogram"
	"github.com/akerl/voyager/prompt"

	speculate "github.com/akerl/speculate/utils"
)

const (
	accountRegexString = `(\d+)(/(\w+))?`
)

var accountRegex = regexp.MustCompile(accountRegexString)

type hop struct {
	Profile string
	Account string
	Role    string
	Mfa     bool
}

// Travel accepts a role and args and turns them into creds
func Travel(targetRole string, args []string) (speculate.Creds, error) {
	return NamedTravel(targetRole, args, "")
}

// NamedTravel accepts a role, args, and session name and turns them into creds
func NamedTravel(targetRole string, args []string, sessionName string) (speculate.Creds, error) {
	var creds speculate.Creds

	cp := cartogram.Pack{}
	if err := cp.Load(); err != nil {
		return creds, err
	}

	targetAccount, err := cp.Find(args)
	if err != nil {
		return creds, err
	}

	if targetRole == "" {
		roleNames := []string{}
		for k := range targetAccount.Roles {
			roleNames = append(roleNames, k)
		}
		if len(roleNames) == 1 {
			targetRole = roleNames[0]
		} else {
			sort.Strings(roleNames)
			targetRole, err = prompt.PickFromList("Desired Role:", roleNames, "")
			if err != nil {
				return creds, err
			}
		}
	} else {
		if _, ok := targetAccount.Roles[targetRole]; !ok {
			return creds, fmt.Errorf("Provided role not present in account")
		}
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
	accountMatch := accountRegex.FindStringSubmatch(a.Source)
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
