package cartogram

import (
	"regexp"
)

const (
	// roleSourceRegexString matches an account number and role name, /-delimited
	// Per https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_iam-limits.html .
	// role names can contain alphanumeric characters, and these symbols: +=,.@_-
	sourceRegexString = `^(\d{12})/([a-zA-Z0-9+=,.@_-]+)$`
)

var sourceRegex = regexp.MustCompile(sourceRegexString)

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
	Name    string    `json:"name"`
	Mfa     bool      `json:"mfa"`
	Sources SourceSet `json:"sources"`
}

// SourceSet is a list of Sources
type SourceSet []Source

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

// AllProfiles returns all unique profiles found
func (as AccountSet) AllProfiles() []string {
	res := []string{}
	for _, x := range as {
		res = append(res, x.AllProfiles()...)
	}
	return uniqCollect(res)
}

// AllProfiles returns all unique profiles found
func (a Account) AllProfiles() []string {
	res := []string{}
	for _, x := range a.Roles {
		res = append(res, x.AllProfiles()...)
	}
	return uniqCollect(res)
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

// AllProfiles returns all unique profiles found
func (rs RoleSet) AllProfiles() []string {
	res := []string{}
	for _, x := range rs {
		res = append(res, x.AllProfiles()...)
	}
	return uniqCollect(res)
}

// AllProfiles returns all unique profiles found
func (r Role) AllProfiles() []string {
	res := []string{}
	for _, x := range r.Sources {
		res = append(res, x.AllProfiles()...)
	}
	return uniqCollect(res)
}

// AllProfiles returns all unique profiles found
func (ss SourceSet) AllProfiles() []string {
	res := []string{}
	for _, x := range ss {
		res = append(res, x.AllProfiles()...)
	}
	return uniqCollect(res)
}

// AllProfiles returns all unique profiles found
func (s Source) AllProfiles() []string {
	if s.IsProfile() {
		return []string{s.Path}
	}
	return []string{}
}

// IsProfile returns true if the source hop is
func (s Source) IsProfile() bool {
	return !sourceRegex.MatchString(s.Path)
}

// Parse returns the account and role for a non-profile Path, or two empty strings
func (s Source) Parse() (string, string) {
	if s.IsProfile() {
		return "", ""
	}
	match := sourceRegex.FindStringSubmatch(s.Path)
	return match[1], match[2]
}
