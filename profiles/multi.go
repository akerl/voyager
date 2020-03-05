package profiles

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

// MultiStore is a storage backend which tries a series of backends
type MultiStore struct {
	Backends []Store
}

// Lookup looks up creds from the list of backends
func (m *MultiStore) Lookup(profile string) (credentials.Value, error) {
	logger.InfoMsgf("looking up %s in multi store", profile)

	var err error
	var creds credentials.Value
	var readIndex int

	for index, item := range m.Backends {
		creds, err = item.Lookup(profile)
		if err == nil {
			readIndex = index
			break
		}
		logger.DebugMsgf("backend failed with error: %s", err)
	}
	if creds.AccessKeyID == "" {
		return credentials.Value{}, fmt.Errorf("all backends failed to return creds")
	}

	writeIndex, writer := m.getWriter()
	if writer != nil && writeIndex < readIndex {
		logger.InfoMsg("writing forward to earlier backend")
		err := writer.Write(profile, creds)
		if err != nil {
			return credentials.Value{}, err
		}
	}

	return creds, nil
}

// Check returns true if any backend has the credentials cached
func (m *MultiStore) Check(profile string) bool {
	logger.InfoMsgf("checking for %s in multi store", profile)
	for _, item := range m.Backends {
		if item.Check(profile) {
			return true
		}
	}
	return false
}

// Delete removes a profile from all backends
func (m *MultiStore) Delete(profile string) error {
	logger.InfoMsgf("deleting %s from multi store", profile)
	for _, item := range m.Backends {
		err := item.Delete(profile)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiStore) Write(s string, c credentials.Value) error {
	_, writer := m.getWriter()
	if writer == nil {
		return fmt.Errorf("no writers available in backends")
	}
	return writer.Write(s, c)
}

func (m *MultiStore) getWriter() (int, WritableStore) {
	logger.InfoMsgf("looking up writer in backends")
	for index, item := range m.Backends {
		writer, ok := item.(WritableStore)
		if ok {
			logger.InfoMsgf("found writer in %d backend", index)
			return index, writer
		}
	}
	logger.InfoMsgf("no writer found")
	return 0, nil
}

// WritableStore defines a backend which can save credentials
type WritableStore interface {
	Write(string, credentials.Value) error
}
