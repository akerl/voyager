package profiles

import (
	"github.com/akerl/timber/v2/log"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

var logger = log.NewLogger("voyager")

// Store is an object which can look up credentials
type Store interface {
	Lookup(string) (credentials.Value, error)
	Check(string) bool
	Delete(string) error
}

// NewDefaultStore returns the default backend set
func NewDefaultStore() Store {
	logger.InfoMsg("initializing the default profiles store")
	return &MultiStore{
		Backends: []Store{
			&KeyringStore{},
			&PromptStore{},
		},
	}
}

// BulkCheck checks which of a slice of profiles exist in the store
func BulkCheck(s Store, all []string) []string {
	found := []string{}
	for _, item := range all {
		if s.Check(item) {
			found = append(found, item)
		}
	}
	return found
}
