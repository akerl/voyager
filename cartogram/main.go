package cartogram

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

const (
	configName = ".cartograms"
)

// Role holds information about authenticating to a role
type Role struct {
	Mfa bool `json:"mfa"`
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

// Lookup finds an account in a Pack based on its ID
func (cp Pack) Lookup(accountId string) (Account, Cartogram, error) {
	for name, c := range cp {
		account, err := c.Lookup(accountId)
	}
}

// Lookup finds an account in a Cartogram based on its ID
func (c Cartogram) Lookup(accountId string) (Account, error) {
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
	var results Cartogram
	if err := json.Unmarshal(data, &results); err != nil {
		return err
	}
	*c = append(*c, results...)
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
	buffer, err := json.MarshalIndent(c, "", "  ")
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
