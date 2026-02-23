package travel

import (
	"fmt"

	"github.com/akerl/voyager/v3/cartogram"
	"github.com/akerl/voyager/v3/pkgver"
	"github.com/akerl/voyager/v3/profiles"

	"github.com/BurntSushi/locker"
	"github.com/akerl/speculate/v2/creds"
)

var mutex *locker.Locker

func init() {
	mutex = locker.NewLocker()
}

// Path defines a set of hops to reach the target account
type Path []Hop

// Hop defines an individual node on the path from initial credentials
// to the target role
type Hop struct {
	Profile string
	Account cartogram.Account
	Role    string
	Mfa     bool
}

// TraverseOptions defines the parameters for traversing a path
type TraverseOptions struct {
	MfaCode        string
	MfaPrompt      creds.MfaPrompt
	Store          profiles.Store
	Cache          Cache
	SessionName    string
	Lifetime       int64
	UserAgentItems []creds.UserAgentItem
}

// DefaultTraverseOptions returns a standard set of TraverseOptions
func DefaultTraverseOptions() TraverseOptions {
	return TraverseOptions{
		MfaPrompt: &creds.DefaultMfaPrompt{},
		Store:     profiles.NewDefaultStore(),
		Cache:     &MapCache{},
	}
}

// Traverse executes a path and returns the final resulting credentials
// using the default set of TraverseOptions
func (p Path) Traverse() (creds.Creds, error) {
	return p.TraverseWithOptions(DefaultTraverseOptions())
}

// TraverseWithOptions executes a path and returns the final resulting credentials
func (p Path) TraverseWithOptions(opts TraverseOptions) (creds.Creds, error) {
	logger.InfoMsgf("traversing path %+v with options %+v", p, opts)

	err := clearEnvironment()
	if err != nil {
		return creds.Creds{}, err
	}

	profileHop, stack := p[0], p[1:]
	logger.InfoMsgf("loading origin hop: %+v", profileHop)
	profileCreds, err := opts.Store.Lookup(profileHop.Profile)
	if err != nil {
		return creds.Creds{}, err
	}

	uai := []creds.UserAgentItem{{
		Name:    "voyager",
		Version: pkgver.Version,
	}}
	for _, x := range opts.UserAgentItems {
		uai = append(uai, x)
	}

	c := creds.Creds{
		AccessKey:      profileCreds.AccessKeyID,
		SecretKey:      profileCreds.SecretAccessKey,
		UserAgentItems: uai,
	}

	for _, thisHop := range stack {
		c, err = thisHop.Traverse(c, opts)
		if err != nil {
			break
		}
	}
	return c, err
}

// Traverse executes a Hop, returning the new credentials
func (h Hop) Traverse(c creds.Creds, opts TraverseOptions) (creds.Creds, error) {
	key := h.toKey()
	mutex.Lock(key)
	defer mutex.Unlock(key)

	if cached, ok := CheckCache(opts.Cache, h); ok {
		return cached, nil
	}
	logger.InfoMsgf("Executing hop: %+v", h)
	a := creds.AssumeRoleOptions{
		RoleName:    h.Role,
		AccountID:   h.Account.Account,
		SessionName: opts.SessionName,
		Lifetime:    opts.Lifetime,
	}

	if h.Mfa {
		a.UseMfa = true
		a.MfaCode = opts.MfaCode
		a.MfaPrompt = opts.MfaPrompt
	}

	c.Region = h.Account.Region
	if c.Region == "" {
		logger.InfoMsg("missing region for hop; inferring us-east-1")
		c.Region = "us-east-1"
	}
	newCreds, err := c.AssumeRole(a)
	if err != nil {
		return creds.Creds{}, err
	}
	err = opts.Cache.Put(h, newCreds)
	return newCreds, err
}

func (h *Hop) toKey() string {
	if h.Profile != "" {
		return fmt.Sprintf("profile--%s", h.Profile)
	}
	return fmt.Sprintf("%s-%s-%t", h.Account.Account, h.Role, h.Mfa)
}
