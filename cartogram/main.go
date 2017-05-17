package cartogram

import (
	"encoding/json"
	"io/ioutil"
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

// Cartograms defines a slice of cartograms
type Cartograms map[string][]Cartogram

// Cartogram defines the spec for a role assumption target
type Cartogram struct {
	Account string            `json:"account"`
	Region  string            `json:"region"`
	Source  string            `json:"source"`
	Roles   map[string]Role   `json:"roles"`
	Tags    map[string]string `json:"tags"`
}

// Load populates the Cartograms from disk
func (c *Cartograms) Load() error {
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
	err = c.loadFromFiles(files)
	return err
}

func (c *Cartograms) loadFromFiles(filePaths []string) error {
	for _, filePath := range filePaths {
		if err := c.loadFromFile(filePath); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cartograms) loadFromFile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	name := path.Base(filePath)
	return c.loadFromString(name, data)
}

func (c *Cartograms) loadFromString(name string, data []byte) error {
	var results []Cartogram
	if err := json.Unmarshal(data, &results); err != nil {
		return err
	}
	(*c)[name] = results
	return nil
}

// Write dumps the Cartograms to disk
func (c *Cartograms) Write() error {
	for name := range *c {
		if err := c.writeToFile(name); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cartograms) writeToFile(name string) error {
	config, err := configDir()
	if err != nil {
		return err
	}
	filePath := path.Join(config, name)
	data, err := c.writeToString(name)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, data, 0600)
	return err
}

func (c *Cartograms) writeToString(name string) ([]byte, error) {
	buffer, err := json.MarshalIndent((*c)[name], "", "  ")
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
	return path.Join(home, configName), nil
}

func homeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}
