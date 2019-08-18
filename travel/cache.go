package travel

import (
	"fmt"
	"github.com/akerl/speculate/v2/creds"
	"github.com/aws/aws-sdk-go/service/sts"
)

// Cache defines a credential caching object
type Cache interface {
	Get(Hop) (creds.Creds, bool)
	Put(Hop, creds.Creds) error
	Delete(Hop) error
}

// CheckCache returns credentials if they exist in the cache and are still valid
// If the credentials exist but are invalid/expired, it removes them from the cache
func CheckCache(c Cache, h Hop) (creds.Creds, bool) {
	logger.DebugMsgf("checking cache for %+v", h)
	cachedCreds, ok := c.Get(h)
	if !ok {
		return creds.Creds{}, false
	}
	client, err := cachedCreds.Client()
	if err == nil {
		_, err := client.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err == nil {
			return cachedCreds, true
		}
	}
	c.Delete(h)
	return creds.Creds{}, false
}

// NullCache implements an empty cache which stores nothing
type NullCache struct{}

// Put is a no-op for NullCache
func (nc *NullCache) Put(_ Hop, _ creds.Creds) error {
	return nil
}

// Get for NullCache always returns empty Creds / false
func (nc *NullCache) Get(_ Hop) (creds.Creds, bool) {
	return creds.Creds{}, false
}

// Delete is a no-op for NullCache
func (nc *NullCache) Delete(_ Hop) error {
	return nil
}

// MapCache stores credentials in a map object based on the hop information
type MapCache struct {
	creds map[string]creds.Creds
}

// Put stores the credentials in the map
func (mc *MapCache) Put(h Hop, c creds.Creds) error {
	key := mc.hopToKey(h)
	logger.DebugMsgf("mapcache: caching %s", key)
	if mc.creds == nil {
		mc.creds = map[string]creds.Creds{}
	}
	mc.creds[key] = c
	return nil
}

// Get returns credentials from the map, if they exist
func (mc *MapCache) Get(h Hop) (creds.Creds, bool) {
	key := mc.hopToKey(h)
	logger.DebugMsgf("mapcache: getting %s", key)
	creds, ok := mc.creds[key]
	return creds, ok
}

// Delete removes credentials from the cache
func (mc *MapCache) Delete(h Hop) error {
	key := mc.hopToKey(h)
	logger.DebugMsgf("mapcache: deleting %s", key)
	delete(mc.creds, key)
	return nil
}

func (mc *MapCache) hopToKey(h Hop) string {
	if h.Profile != "" {
		return fmt.Sprintf("profile--%s", h.Profile)
	}
	return fmt.Sprintf("%s-%s-%t", h.Account, h.Role, h.Mfa)
}
