package utils

import (
	"fmt"
	"os"
	"regexp"
	"sort"

	"github.com/akerl/voyager/cartogram"

	speculate "github.com/akerl/speculate/utils"
	"github.com/dixonwille/wmenu"
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
	var creds speculate.Creds

	cp := cartogram.Pack{}
	if err := cp.Load(); err != nil {
		return creds, err
	}

	targetAccount, err := findAccount(cp, args)
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
			targetRole, err = pickFromList("Desired Role:", roleNames, "")
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
			RoleName:  thisHop.Role,
			AccountID: thisHop.Account,
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

func findAccount(cp cartogram.Pack, args []string) (cartogram.Account, error) {
	var targetAccount cartogram.Account
	var err error
	var found bool

	found, targetAccount, err = findDirectAccount(cp, args)
	if err != nil || found {
		return targetAccount, err
	}

	found, targetAccount, err = findMatchAccount(cp, args)
	if err != nil || found {
		return targetAccount, err
	}

	return targetAccount, fmt.Errorf("Unable to locate an account with provided info")
}

func findDirectAccount(cp cartogram.Pack, args []string) (bool, cartogram.Account, error) {
	var account cartogram.Account
	if len(args) != 1 {
		return false, account, nil
	}
	accountMatch := accountRegex.FindStringSubmatch(args[0])
	if len(accountMatch) == 0 {
		return false, account, nil
	}
	var accountID string
	accountID = accountMatch[1]
	found, account := cp.Lookup(accountID)
	if !found {
		return false, account, fmt.Errorf("Account not found: %s", accountID)
	}
	return true, account, nil
}

func findMatchAccount(cp cartogram.Pack, args []string) (bool, cartogram.Account, error) {
	var account cartogram.Account
	tfs := cartogram.TagFilterSet{}
	if err := tfs.LoadFromArgs(args); err != nil {
		return false, account, err
	}
	accounts := cp.Search(tfs)

	switch len(accounts) {
	case 0:
		return false, account, nil
	case 1:
		return true, accounts[0], nil
	default:
		mapOfAccounts := map[string]cartogram.Account{}
		sliceOfNames := []string{}
		for _, a := range accounts {
			name := fmt.Sprintf("%s (%s)", a.Account, a.Tags)
			mapOfAccounts[name] = a
			sliceOfNames = append(sliceOfNames, name)
		}
		chosen, err := pickFromList("Desired account:", sliceOfNames, "")
		if err != nil {
			return false, account, err
		}
		return true, mapOfAccounts[chosen], nil
	}
}

func pickFromList(message string, list []string, defaultOpt string) (string, error) {
	c := make(chan string, 1)

	menu := wmenu.NewMenu(message)
	menu.ChangeReaderWriter(os.Stdin, os.Stderr, os.Stderr)
	menu.LoopOnInvalid()
	menu.Action(func(opts []wmenu.Opt) error {
		c <- opts[0].Value.(string)
		return nil
	})

	for _, item := range list {
		isDefault := false
		if item == defaultOpt {
			isDefault = true
		}
		menu.Option(item, item, isDefault, nil)
	}

	if err := menu.Run(); err != nil {
		return "", err
	}

	return <-c, nil
}
