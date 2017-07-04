package cartogram

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"

	"github.com/akerl/voyager/prompt"
)

const (
	accountRegexString = `(\d+)(/(\w+))?`
)

var accountRegex = regexp.MustCompile(accountRegexString)

// Pack defines a group of Cartograms
type Pack map[string]Cartogram

// Find checks both Lookup and Search for an account
func (cp Pack) Find(args []string) (Account, error) {
	var targetAccount Account
	var err error
	var found bool

	found, targetAccount, err = cp.findDirectAccount(args)
	if err != nil || found {
		return targetAccount, err
	}

	found, targetAccount, err = cp.findMatchAccount(args)
	if err != nil || found {
		return targetAccount, err
	}

	return targetAccount, fmt.Errorf("Unable to locate an account with provided info")
}

func (cp Pack) findDirectAccount(args []string) (bool, Account, error) {
	var account Account
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

func (cp Pack) findMatchAccount(args []string) (bool, Account, error) {
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
		mapOfAccounts := map[string]Account{}
		sliceOfNames := []string{}
		for _, a := range accounts {
			name := fmt.Sprintf("%s (%s)", a.Account, a.Tags)
			mapOfAccounts[name] = a
			sliceOfNames = append(sliceOfNames, name)
		}
		chosen, err := prompt.PickFromList("Desired account:", sliceOfNames, "")
		if err != nil {
			return false, account, err
		}
		return true, mapOfAccounts[chosen], nil
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
	var files []string
	for _, fileObj := range fileObjs {
		files = append(files, path.Join(config, fileObj.Name()))
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
