package cartogram

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"regexp"
	"strings"
)

const (
	configName  = ".cartograms"
	specVersion = 0
)

type versionedCartogram struct {
	Version   int       `json:"version"`
	Cartogram Cartogram `json:"cartogram"`
}

// Pack defines a group of Cartograms
type Pack map[string]Cartogram

// Cartogram defines a set of Accounts
type Cartogram []Account

// Account defines the spec for a role assumption target
type Account struct {
	Account string            `json:"account"`
	Region  string            `json:"region"`
	Source  string            `json:"source"`
	Roles   map[string]Role   `json:"roles"`
	Tags    map[string]string `json:"tags"`
}

// Role holds information about authenticating to a role
type Role struct {
	Mfa bool `json:"mfa"`
}

// TagFilter describes a filter to apply based on an account's tags
type TagFilter struct {
	Name  string
	Value *regexp.Regexp
}

// TagFilterSet describes a set of tag filters
type TagFilterSet []TagFilter

// LoadFromArgs parses key:value args into a TagFilterSet
func (tfs *TagFilterSet) LoadFromArgs(args []string) error {
	var err error
	for _, a := range args {
		tf := TagFilter{}
		fields := strings.SplitN(a, ":", 2)
		var regexString string
		if len(fields) == 1 {
			regexString = fields[0]
		} else {
			tf.Name = fields[0]
			regexString = fields[1]
		}
		tf.Value, err = regexp.Compile(regexString)
		if err != nil {
			return err
		}
		*tfs = append(*tfs, tf)
	}
	return nil
}

// Match checks if an account matches the tag filter
func (tf TagFilter) Match(a Account) bool {
	for tagName, tagValue := range a.Tags {
		// TODO: break this out to check tagName if set
		if tf.Name == "" || tf.Name == tagName {
			if tf.Value.MatchString(tagValue) {
				return true
			}
		}
	}
	return false
}

// Match checks if an account matches the tag filter set
func (tfs TagFilterSet) Match(a Account) bool {
	for _, tf := range tfs {
		if !tf.Match(a) {
			return false
		}
	}
	return true
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

// Lookup finds an account in a Cartogram based on its ID
func (c Cartogram) Lookup(accountID string) (bool, Account) {
	for _, a := range c {
		if a.Account == accountID {
			return true, a
		}
	}
	return false, Account{}
}

// Search finds accounts based on their tags
func (cp Pack) Search(tfs TagFilterSet) Cartogram {
	results := Cartogram{}
	for _, c := range cp {
		results = append(results, c.Search(tfs)...)
	}
	return results
}

// Search finds accounts based on their tags
func (c Cartogram) Search(tfs TagFilterSet) Cartogram {
	results := Cartogram{}
	for _, a := range c {
		if tfs.Match(a) {
			results = append(results, a)
		}
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

func (c *Cartogram) loadFromFile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return c.loadFromString(data)
}

func (c *Cartogram) loadFromString(data []byte) error {
	var results versionedCartogram
	if err := json.Unmarshal(data, &results); err != nil {
		return err
	}
	if results.Version != specVersion {
		return fmt.Errorf("Spec version mismatch: expected %d, got %d", specVersion, results.Version)
	}
	*c = results.Cartogram
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

func (c Cartogram) writeToFile(filePath string) error {
	data, err := c.writeToString()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, data, 0600)
	return err
}

func (c Cartogram) writeToString() ([]byte, error) {
	vc := versionedCartogram{Version: specVersion, Cartogram: c}
	buffer, err := json.MarshalIndent(vc, "", "  ")
	if err != nil {
		return []byte{}, err
	}
	return buffer, nil
}

func configDir() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", err
	}
	dir := path.Join(home, configName)
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func homeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}
