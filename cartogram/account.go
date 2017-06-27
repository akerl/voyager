package cartogram

import ()

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
