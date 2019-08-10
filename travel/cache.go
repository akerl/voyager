package travel

import (
	"github.com/akerl/speculate/v2/creds"
)

type Cache struct {
	creds map[string]creds.Creds
}

func (cache *Cache) Put(h Hop, c creds.Creds) error {
	// TODO: write
	return nil
}

func (cache *Cache) Get(h Hop) (creds.Creds, bool) {
	// TODO: write
	return creds.Creds{}, false
}

func (cache *Cache) hopToKey(h Hop) string {
	// TODO: write
	return ""
}
