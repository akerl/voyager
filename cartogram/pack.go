package cartogram

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"

	"github.com/akerl/input/list"
)

const (
	// accountRegexString matches an account number with an optional /$role_name
	// Per https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_iam-limits.html .
	// role names can contain alphanumeric characters, and these symbols: +=,.@_-
	accountRegexString = `^(\d{12})$`
)

var accountRegex = regexp.MustCompile(accountRegexString)

// Pack defines a group of Cartograms
type Pack map[string]Cartogram

// Find checks both Lookup and Search for an account
func (cp Pack) Find(args []string) (Account, error) {
	return cp.FindWithPrompt(args, list.Default())
}

// FindWithPrompt checks both Lookup and Search for an account with a custom prompt
func (cp Pack) FindWithPrompt(args []string, prompt list.Prompt) (Account, error) {
	var targetAccount Account
	var err error
	var found bool

	found, targetAccount, err = cp.findDirectAccount(args)
	if err != nil || found {
		return targetAccount, err
	}

	found, targetAccount, err = cp.findMatchAccount(args, prompt)
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
	accountID := args[0]
	logger.InfoMsgf("looking up account as potential direct match: %s", accountID)
	if !accountRegex.MatchString(accountID) {
		return false, account, nil
	}
	logger.InfoMsg("input matched as an account id")
	found, account := cp.Lookup(accountID)
	if !found {
		return false, account, fmt.Errorf("account not found: %s", accountID)
	}
	return true, account, nil
}

func (cp Pack) findMatchAccount(args []string, prompt list.Prompt) (bool, Account, error) {
	logger.InfoMsgf("looking for matching account using provided args: %v", args)

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
		logger.InfoMsgf("found single matching account: %s", accounts[0].Account)
		return true, accounts[0], nil
	default:
		logger.InfoMsgf("found %d matches", len(accounts))
		optSet := make(list.OptionSet, len(accounts))
		for index, account := range accounts {
			optSet[index] = list.Option{Name: account.Account, Metadata: account.Tags}
		}
		index, err := prompt.Execute("Pick an account", optSet)
		if err != nil {
			return false, account, err
		}
		return true, accounts[index], nil
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

func (cp Pack) toSlice() []Cartogram {
	result := []Cartogram{}
	for _, v := range cp {
		result = append(result, v)
	}
	return result
}

// AllProfiles returns all unique profiles found
func (cp Pack) AllProfiles() []string {
	res := []string{}
	for _, x := range cp {
		res = append(res, x.AllProfiles()...)
	}
	return uniqCollect(res)
}

// Load populates the Cartograms from disk
func (cp Pack) Load() error {
	logger.InfoMsg("loading pack from disk")
	config, err := configDir()
	if err != nil {
		return err
	}
	fileObjs, err := ioutil.ReadDir(config)
	if err != nil {
		return err
	}
	logger.InfoMsgf("found %d cartogram files", len(fileObjs))
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
	logger.InfoMsg("writing pack to disk")
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
