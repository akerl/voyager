package cartogram

import (
	"io/ioutil"
	"path"

	"github.com/akerl/voyager/utils"
)

// Pack defines a group of Cartograms
type Pack map[string]Cartogram

// Find checks both Lookup and Search for an account
func (cp Pack) Find(args []string) (Account, error) {
	return findAccount(cp, args)
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
