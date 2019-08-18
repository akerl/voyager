package travel

import (
	"github.com/akerl/voyager/v2/profiles"

	"github.com/akerl/speculate/v2/creds"
)

// Path defines a set of hops to reach the target account
type Path []Hop

// Hop defines an individual node on the path from initial credentials
// to the target role
type Hop struct {
	Profile string
	Account string
	Role    string
	Mfa     bool
	Region  string
}

// TraverseOptions defines the parameters for traversing a path
type TraverseOptions struct {
	MfaCode     string
	MfaPrompt   creds.MfaPrompt
	Store       profiles.Store
	Cache       Cache
	SessionName string
	Lifetime    int64
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
	c := creds.Creds{
		AccessKey: profileCreds.AccessKeyID,
		SecretKey: profileCreds.SecretAccessKey,
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
	if cached, ok := CheckCache(opts.Cache, h); ok {
		return cached, nil
	}
	logger.InfoMsgf("Executing hop: %+v", h)
	a := creds.AssumeRoleOptions{
		RoleName:    h.Role,
		AccountID:   h.Account,
		SessionName: opts.SessionName,
		Lifetime:    opts.Lifetime,
	}

	if h.Mfa {
		a.UseMfa = true
		a.MfaCode = opts.MfaCode
		a.MfaPrompt = opts.MfaPrompt
	}

	c.Region = h.Region
	newCreds, err := c.AssumeRole(a)
	if err != nil {
		return creds.Creds{}, err
	}
	err = opts.Cache.Put(h, newCreds)
	return newCreds, err
}
