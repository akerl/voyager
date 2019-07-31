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
	Account string `json:"account"`
	Region  string `json:"region"`
	Roles   []Role `json:"roles"`
	Tags    Tags   `json:"tags"`
}

// Role holds information about authenticating to a role
type Role struct {
	Name    string   `json:"name"`
	Sources []Source `json:"sources"`
}

// Source defines the previous hop for accessing a role
type Source struct {
	Mfa  bool   `json:"mfa"`
	Path string `json:"path"`
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

func (a Account) rolesAsMap() map[string]Role {
	roles := make(map[string]Role, len(a.Roles))
	for _, i := range a.Roles {
		roles[i.Name] = i
	}
	return roles
}

func (a Account) roleNames() []string {
	names := make([]string, len(a.Roles))
	for index, item := range a.Roles {
		names[index] = item.Name
	}
	return names
}

// PickRole returns a role from the account
func (a Account) PickRole(roleName string) (Role, error) {
	return a.PickRoleWithPrompt(roleName, prompt.WithDefault)
}

// PickRoleWithPrompt returns a role from the account with a custom prompt
func (a Account) PickRoleWithPrompt(roleName string, pf prompt.Func) (Role, error) {
	rolesAsMap := a.rolesAsMap()
	if roleName != "" {
		r, ok := rolesAsMap[roleName]
		if !ok {
			return Role{}, fmt.Errorf("provided role not present in account")
		}
		return r, nil
	}

	if len(a.Roles) == 1 {
		return a.Roles[0], nil
	}

	roleNames := a.roleNames()
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
		return Role{}, err
	}

	return rolesAsMap[roleNames[index]], nil
}
