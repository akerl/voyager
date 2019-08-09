package cartogram

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// Cartogram defines a set of accounts and their metadata
type Cartogram struct {
	Version    int        `json:"version"`
	Created    time.Time  `json:"created"`
	AccountSet AccountSet `json:"accounts"`
}

// dummyCartogram just parses the Version
// this is used to test for schema version mismatch
type dummyCartogram struct {
	Version int `json:"version"`
}

// NewCartogram creates a new cartogram from an account set
func NewCartogram(as AccountSet) Cartogram {
	return Cartogram{
		Version:    specVersion,
		Created:    time.Now(),
		AccountSet: as,
	}
}

// Lookup finds an account in a Cartogram based on its ID
func (c Cartogram) Lookup(accountID string) (bool, Account) {
	return c.AccountSet.Lookup(accountID)
}

// Search finds accounts based on their tags
func (c Cartogram) Search(tfs TagFilterSet) AccountSet {
	return c.AccountSet.Search(tfs)
}

func (c *Cartogram) loadFromFile(filePath string) error {
	logger.InfoMsgf("loading cartogram from %s", filePath)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return c.loadFromString(data)
}

func (c *Cartogram) loadFromString(data []byte) error {
	if err := schemaVersionCheck(data); err != nil {
		return err
	}
	return json.Unmarshal(data, &c)
}

func schemaVersionCheck(data []byte) error {
	var c Cartogram
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}
	if c.Version != specVersion {
		return SpecVersionError{ActualVersion: c.Version, ExpectedVersion: specVersion}
	}
	return nil
}

func (c Cartogram) writeToFile(filePath string) error {
	logger.InfoMsgf("writing cartogram to %s", filePath)
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
