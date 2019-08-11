package cartogram

// AccountSet is a set of accounts
type AccountSet []Account

// Account defines the spec for a role assumption target
type Account struct {
	Account string  `json:"account"`
	Region  string  `json:"region"`
	Roles   RoleSet `json:"roles"`
	Tags    Tags    `json:"tags"`
}

// RoleSet is a list of Roles
type RoleSet []Role

// Role holds information about authenticating to a role
type Role struct {
	Name    string   `json:"name"`
	Mfa     bool     `json:"mfa"`
	Sources []Source `json:"sources"`
}

// Source defines the previous hop for accessing a role
type Source struct {
	Path string `json:"path"`
}

// Lookup finds an account in a Cartogram based on its ID
func (as AccountSet) Lookup(accountID string) (bool, Account) {
	logger.InfoMsgf("looking up accountID %s in set", accountID)
	for _, a := range as {
		if a.Account == accountID {
			return true, a
		}
	}
	return false, Account{}
}

// Search finds accounts based on their tags
func (as AccountSet) Search(tfs TagFilterSet) AccountSet {
	logger.InfoMsgf("searching for %v in set", tfs)
	results := AccountSet{}
	for _, a := range as {
		if tfs.Match(a) {
			results = append(results, a)
		}
	}
	return results
}

// Lookup searches for a role by name
func (rs RoleSet) Lookup(name string) (bool, Role) {
	logger.InfoMsgf("looking up role %s in set", name)
	for _, r := range rs {
		if r.Name == name {
			return true, r
		}
	}
	return false, Role{}
}
