package cmd

import (
	"fmt"
	"os"
	"regexp"
	"sort"

	"github.com/akerl/voyager/cartogram"

	"github.com/dixonwille/wmenu"
	"github.com/spf13/cobra"
)

const (
	accountRegexString = `(\d+)(/(\w+))?`
)

var accountRegex = regexp.MustCompile(accountRegexString)

var travelCmd = &cobra.Command{
	Use:   "travel",
	Short: "Resolve creds for a AWS account",
	RunE:  travelRunner,
}

func init() {
	rootCmd.AddCommand(travelCmd)
}

func travelRunner(cmd *cobra.Command, args []string) error {
	cp := cartogram.Pack{}
	if err := cp.Load(); err != nil {
		return err
	}

	targetAccount, targetRole, err := findAccount(cp, args)
	if err != nil {
		return err
	}

	if targetRole == "" {
		roleNames := []string{}
		for k := range targetAccount.Roles {
			roleNames = append(roleNames, k)
		}
		sort.Strings(roleNames)
		targetRole, err = pickFromList("Desired Role:", roleNames, "")
		if err != nil {
			return err
		}
	}

	fmt.Printf("Account: %s\nRole: %s\n", targetAccount, targetRole)
	return nil
}

func findAccount(cp cartogram.Pack, args []string) (cartogram.Account, string, error) {
	var targetAccount cartogram.Account
	var targetRole string
	var err error
	var found bool

	found, targetAccount, targetRole, err = findDirectAccount(cp, args)
	if err != nil {
		return targetAccount, "", err
	}
	if found {
		return targetAccount, targetRole, nil
	}

	found, targetAccount, targetRole, err = findMatchAccount(cp, args)
	if found {
		return targetAccount, targetRole, nil
	}

	return targetAccount, targetRole, fmt.Errorf("Unable to locate an account with provided info")
}

func findDirectAccount(cp cartogram.Pack, args []string) (bool, cartogram.Account, string, error) {
	var account cartogram.Account
	if len(args) != 1 {
		return false, account, "", nil
	}
	accountMatch := accountRegex.FindStringSubmatch(args[0])
	if len(accountMatch) == 0 {
		return false, account, "", nil
	}
	var accountID, role string
	accountID = accountMatch[1]
	if len(accountMatch) > 2 {
		role = accountMatch[3]
	}
	found, account := cp.Lookup(accountID)
	if !found {
		return false, account, role, fmt.Errorf("Account not found: %s", accountID)
	}
	if role != "" {
		if _, ok := account.Roles[role]; !ok {
			return false, account, role, fmt.Errorf("Role not present in account")
		}
	}
	return true, account, role, nil
}

func findMatchAccount(cp cartogram.Pack, args []string) (bool, cartogram.Account, string, error) {
	var account cartogram.Account
	tfs := cartogram.TagFilterSet{}
	if err := tfs.LoadFromArgs(args); err != nil {
		return false, account, "", err
	}
	accounts := cp.Search(tfs)

	switch len(accounts) {
	case 0:
		return false, account, "", nil
	case 1:
		return true, accounts[0], "", nil
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
			return false, account, "", err
		}
		return true, mapOfAccounts[chosen], "", nil
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
