package multi

import (
	"strings"

	"github.com/akerl/voyager/v2/travel"

	"github.com/akerl/speculate/v2/creds"
)

// Processor defines the settings for parallel processing
type Processor struct {
	Grapher      travel.Grapher
	Options      travel.TraverseOptions
	Args         []string
	RoleNames    []string
	ProfileNames []string
}

func (p Processor) ExecString(cmd string) (map[string]creds.ExecResult, error) {
	cmdSlice := strings.Split(cmd, " ")
	return p.Exec(cmdSlice)
}

func (p Processor) Exec(cmd []string) (map[string]creds.ExecResult, error) {
	paths, err := p.Grapher.ResolveAll(p.Args, p.RoleNames, p.ProfileNames)
	if err != nil {
		return map[string]creds.ExecResult{}, err
	}

	// TODO: Show progress bar
	// TODO: Parallelize this
	// TODO: Return partial cred errors cleanly

	allCreds := map[string]creds.Creds{}
	for _, item := range paths {
		c, err := item.TraverseWithOptions(p.Options)
		if err != nil {
			return map[string]creds.ExecResult{}, err
		}
		accountId, err := c.AccountID()
		if err != nil {
			return map[string]creds.ExecResult{}, err
		}
		allCreds[accountId] = c
	}

	output := map[string]creds.ExecResult{}
	for accountId, c := range allCreds {
		output[accountId] = c.Exec(cmd)
	}
	return output, nil
}
