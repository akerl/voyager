package profiles

import (
	"fmt"
	"os"

	"github.com/akerl/timber/log"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

var logger = log.NewLogger("voyager")

const (
	// EnvVarName defines where to pass the profile name between disconnected functions
	EnvVarName = "VOYAGER_PROFILE"
)

// Store is an object which can look up credentials
type Store interface {
	Lookup(string) (credentials.Value, error)
}

// SetProfile exports variables from a lookup
func SetProfile(profile string, s Store) error {
	creds, err := s.Lookup(profile)
	if err != nil {
		return err
	}

	err = os.Setenv(EnvVarName, profile)
	if err != nil {
		return err
	}

	credsAsMap := map[string]string{
		"AWS_ACCESS_KEY_ID":     creds.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY": creds.SecretAccessKey,
	}
	for k, v := range credsAsMap {
		logger.InfoMsg(fmt.Sprintf("Setting env var: %s", k))
		err = os.Setenv(k, v)
		if err != nil {
			return fmt.Errorf("failed to set env var: %s = %s", k, v)
		}
	}

	return nil
}

// NewDefaultStore returns the default backend set
func NewDefaultStore() Store {
	return &MultiStore{
		Backends: []Store{
			&KeyringStore{},
			&PromptStore{},
		},
	}
}
