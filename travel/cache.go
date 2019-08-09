package travel

import (
	"github.com/akerl/speculate/v2/creds"
)

type Cache struct {
	creds map[string]creds.Creds
}

func (c *Cache) Put(h Hop, c creds.Creds) error {
	// TODO: write
}

func (c *Cache) Get(h Hop) (creds.Creds, bool) {
	// TODO: write
}

func (c *Cache) hopToKey(h Hop) string {
	// TODO: write
}
