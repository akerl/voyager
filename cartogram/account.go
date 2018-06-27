package cartogram

import (
	"fmt"
	"sort"

	"github.com/akerl/voyager/prompt"
)

// AccountSet is a set of accounts
type AccountSet []Account

// Account defines the spec for a role assumption target
type Account struct {
	Account string          `json:"account"`
	Region  string          `json:"region"`
	Source  string          `json:"source"`
	Roles   map[string]Role `json:"roles"`
	Tags    Tags            `json:"tags"`
}

// Role holds information about authenticating to a role
type Role struct {
	Mfa bool `json:"mfa"`
}

// Lookup finds an account in a Cartogram based on its ID
func (as AccountSet) Lookup(accountID string) (bool, Account) {
	for _, a := range as {
		if a.Account == accountID {
			return true, a
		}
	}
	return false, Account{}
}

// Search finds accounts based on their tags
func (as AccountSet) Search(tfs TagFilterSet) AccountSet {
	results := AccountSet{}
	for _, a := range as {
		if tfs.Match(a) {
			results = append(results, a)
		}
	}
	return results
}

// PickRole returns a role from the account
func (a Account) PickRole(roleName string) (string, error) {
	return a.PickRoleWithPrompt(roleName, prompt.WithDefault)
}

// PickRoleWithPrompt returns a role from the account with a custom prompt
func (a Account) PickRoleWithPrompt(roleName string, pf prompt.Func) (string, error) {
	if roleName != "" {
		if _, ok := a.Roles[roleName]; !ok {
			return "", fmt.Errorf("Provided role not present in account")
		}
		return roleName, nil
	}
	roleNames := []string{}
	for k := range a.Roles {
		roleNames = append(roleNames, k)
	}
	if len(roleNames) == 1 {
		return roleNames[0], nil
	}
	sort.Strings(roleNames)
	roleSlices := [][]string{}
	for _, k := range roleNames {
		roleSlices = append(roleSlices, []string{k})
	}

	pa := prompt.Args{
		Message: "Desired Role:",
		Options: roleSlices,
	}
	index, err := pf(pa)
	if err != nil {
		return "", err
	}

	return roleNames[index], nil
}
