package travel

import (
	"fmt"
	"github.com/akerl/speculate/v2/creds"
	"github.com/aws/aws-sdk-go/service/sts"
)

type Cache interface {
	Get(Hop) (creds.Creds, bool)
	Put(Hop, creds.Creds) error
	Delete(Hop) error
}

func CheckCache(c Cache, h Hop) (creds.Creds, bool) {
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

type NullCache struct{}

func (nc *NullCache) Put(_ Hop, _ creds.Creds) error {
	return nil
}

func (nc *NullCache) Get(_ Hop) (creds.Creds, bool) {
	return creds.Creds{}, false
}

func (nc *NullCache) Delete(_ Hop) error {
	return nil
}

type MapCache struct {
	creds map[string]creds.Creds
}

func (mc *MapCache) Put(h Hop, c creds.Creds) error {
	key := mc.hopToKey(h)
	if mc.creds == nil {
		mc.creds = map[string]creds.Creds{}
	}
	mc.creds[key] = c
	return nil
}

func (mc *MapCache) Get(h Hop) (creds.Creds, bool) {
	key := mc.hopToKey(h)
	creds, ok := mc.creds[key]
	return creds, ok
}

func (mc *MapCache) Delete(h Hop) error {
	key := mc.hopToKey(h)
	delete(mc.creds, key)
	return nil
}

func (mc *MapCache) hopToKey(h Hop) string {
	if h.Profile != "" {
		return fmt.Sprintf("profile--%s", h.Profile)
	}
	return fmt.Sprintf("%s-%s-%t", h.Account, h.Role, h.Mfa)
}
