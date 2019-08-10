package travel

import (
	"os"

	"github.com/akerl/speculate/v2/creds"
)

type Path []Hop

type Hop struct {
	Profile string
	Account string
	Role    string
	Mfa     bool
}

type TraverseOptions struct {
	MfaPrompt creds.MfaPrompt
	Cache     *Cache
}

// TODO: clean up all Path.go code

func (p Path) Traverse() (creds.Creds, error) {
	return p.TraverseWithOptions(TraverseOptions{})
}

func clearEnvironment() error {
	for varName := range creds.Translations["envvar"] {
		logger.InfoMsgf("Unsetting env var: %s", varName)
		err := os.Unsetenv(varName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p Path) TraverseWithOptions(opts TraverseOptions) (creds.Creds, error) {
	err := clearEnvironment()
	if err != nil {
		return creds.Creds{}, err
	}

	profileHop, stack := path[0], path[1:]
	logger.InfoMsgf("Setting origin hop: %+v", profileHop)
	store := i.getStore()
	profileCreds, err := store.Lookup(profileHop.Profile)
	if err != nil {
		return creds.Creds{}, err
	}
	// TODO: move this to creds.NewFromValue
	c := creds.Creds{
		AccessKey: profileCreds.AccessKeyID,
		SecretKey: profileCreds.SecretAccessKey,
		Region:    stack[0].Region,
	}

	stack[len(stack)-1].Policy = i.Policy

	for _, thisHop := range stack {
		c, err = i.executeHop(thisHop, c)
		if err != nil {
			break
		}
	}
	return c, err
}

func (h Hop) Traverse(c creds.Creds) (creds.Creds, error) {
	var newCreds creds.Creds
	var err error

	logger.InfoMsgf("Executing hop: %+v", thisHop)
	a := creds.AssumeRoleOptions{
		RoleName:    thisHop.Role,
		AccountID:   thisHop.Account,
		SessionName: i.SessionName,
		Policy:      thisHop.Policy,
		Lifetime:    i.Lifetime,
	}

	if thisHop.Mfa {
		a.UseMfa = true
		a.MfaCode = i.MfaCode
		a.MfaPrompt = i.MfaPrompt
	}
	newCreds, err = c.AssumeRole(a)
	return newCreds, err
}
