package cmd

import (
	"fmt"
	"regexp"

	"github.com/akerl/voyager/cartogram"

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
	return true, account, role, nil
}

func findMatchAccount(cp cartogram.Pack, args []string) (bool, cartogram.Account, string, error) {
	var account cartogram.Account
	tfs := cartogram.TagFilterSet{}
	if err := tfs.LoadFromArgs(args); err != nil {
		return false, account, "", err
	}
	accounts := cp.Search(tfs)
	if len(accounts) == 0 {
		return false, account, "", nil
	}
	return true, accounts[0], "", nil
}
