package travel

import (
	"github.com/akerl/voyager/v2/profiles"

	"github.com/akerl/speculate/v2/creds"
)

type Path []Hop

type Hop struct {
	Profile string
	Account string
	Role    string
	Mfa     bool
	Region  string
}

type TraverseOptions struct {
	MfaCode    string
	MfaPrompt  creds.MfaPrompt
	Store      profiles.Store
	Cache      Cache
	SessioName string
	Lifetime   int64
}

func DefaultTraverseOptions() TraverseOptions {
	return TraverseOptions{
		MfaPrompt: &creds.DefaultMfaPrompt{},
		Store:     profiles.NewDefaultStore(),
		Cache:     &MapCache{},
	}
}

func (p Path) Traverse() (creds.Creds, error) {
	return p.TraverseWithOptions(DefaultTraverseOptions())
}

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

func (h Hop) Traverse(c creds.Creds, opts TraverseOptions) (creds.Creds, error) {
	if cached, ok := CheckCache(opts.Cache, h); ok {
		return cached, nil
	}
	logger.InfoMsgf("Executing hop: %+v", h)
	a := creds.AssumeRoleOptions{
		RoleName:    h.Role,
		AccountID:   h.Account,
		SessionName: opts.SessioName,
		Lifetime:    opts.Lifetime,
	}

	if h.Mfa {
		a.UseMfa = true
		a.MfaCode = opts.MfaCode
		a.MfaPrompt = opts.MfaPrompt
	}

	c.Region = h.Region
	newCreds, err := c.AssumeRole(a)
	return newCreds, err
}
