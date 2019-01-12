package cartogram

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/akerl/voyager/prompt"
)

const (
	// accountRegexString matches an account number with an optional /$role_name
	// Per https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_iam-limits.html .
	// role names can contain alphanumeric characters, and these symbols: +=,.@_-
	accountRegexString = `^(\d+)(?:/([a-zA-Z0-9+=,.@_-]+))?$`
)

// AccountRegex matches an account number with an optional role name
var AccountRegex = regexp.MustCompile(accountRegexString)

// Pack defines a group of Cartograms
type Pack map[string]Cartogram

// Find checks both Lookup and Search for an account
func (cp Pack) Find(args []string) (Account, error) {
	return cp.FindWithPrompt(args, prompt.WithDefault)
}

// FindWithPrompt checks both Lookup and Search for an account with a custom prompt
func (cp Pack) FindWithPrompt(args []string, pf prompt.Func) (Account, error) {
	var targetAccount Account
	var err error
	var found bool

	found, targetAccount, err = cp.findDirectAccount(args)
	if err != nil || found {
		return targetAccount, err
	}

	found, targetAccount, err = cp.findMatchAccount(args, pf)
	if err != nil || found {
		return targetAccount, err
	}

	return targetAccount, fmt.Errorf("unable to locate an account with provided info")
}

func (cp Pack) findDirectAccount(args []string) (bool, Account, error) {
	var account Account
	if len(args) != 1 {
		return false, account, nil
	}
	accountMatch := AccountRegex.FindStringSubmatch(args[0])
	if len(accountMatch) == 0 {
		return false, account, nil
	}
	accountID := accountMatch[1]
	found, account := cp.Lookup(accountID)
	if !found {
		return false, account, fmt.Errorf("account not found: %s", accountID)
	}
	return true, account, nil
}

func (cp Pack) findMatchAccount(args []string, pf prompt.Func) (bool, Account, error) {
	var account Account
	tfs := TagFilterSet{}
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
		sliceOfNames := [][]string{}
		for _, a := range accounts {
			accountSlice := []string{}
			accountSlice = append(accountSlice, a.Account)
			tagSlice := []string{}
			for k, v := range a.Tags {
				tagString := strings.Join([]string{k, v}, ":")
				tagSlice = append(tagSlice, tagString)
			}
			sort.Strings(tagSlice)
			accountSlice = append(accountSlice, tagSlice...)
			sliceOfNames = append(sliceOfNames, accountSlice)
		}
		pa := prompt.Args{
			Message: "Desired Account:",
			Options: sliceOfNames,
		}
		chosen, err := pf(pa)
		if err != nil {
			return false, account, err
		}
		return true, accounts[chosen], nil
	}
}

// Lookup finds an account in a Pack based on its ID
func (cp Pack) Lookup(accountID string) (bool, Account) {
	for _, c := range cp {
		found, account := c.Lookup(accountID)
		if found {
			return true, account
		}
	}
	return false, Account{}
}

// Search finds accounts based on their tags
func (cp Pack) Search(tfs TagFilterSet) AccountSet {
	results := AccountSet{}
	for _, c := range cp {
		results = append(results, c.Search(tfs)...)
	}
	return results
}

// Load populates the Cartograms from disk
func (cp Pack) Load() error {
	config, err := configDir()
	if err != nil {
		return err
	}
	fileObjs, err := ioutil.ReadDir(config)
	if err != nil {
		return nil
	}
	files := make([]string, len(fileObjs))
	for index, fileObj := range fileObjs {
		files[index] = path.Join(config, fileObj.Name())
	}
	err = cp.loadFromFiles(files)
	return err
}

func (cp Pack) loadFromFiles(filePaths []string) error {
	for _, filePath := range filePaths {
		name := path.Base(filePath)
		newC := Cartogram{}
		if err := newC.loadFromFile(filePath); err != nil {
			return err
		}
		cp[name] = newC
	}
	return nil
}

// Write dumps the Cartograms to disk
func (cp Pack) Write() error {
	config, err := configDir()
	if err != nil {
		return err
	}

	for name, c := range cp {
		filePath := path.Join(config, name)
		if err := c.writeToFile(filePath); err != nil {
			return err
		}
	}
	return nil
}
