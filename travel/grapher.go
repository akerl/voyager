package travel

import (
	"regexp"

	"github.com/akerl/voyager/v2/cartogram"
	"github.com/akerl/voyager/v2/profiles"

	"github.com/akerl/input/list"
)

const (
	// roleSourceRegexString matches an account number and role name, /-delimited
	// Per https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_iam-limits.html .
	// role names can contain alphanumeric characters, and these symbols: +=,.@_-
	roleSourceRegexString = `^(\d{12})/([a-zA-Z0-9+=,.@_-]+)$`
)

var roleSourceRegex = regexp.MustCompile(roleSourceRegexString)

type Grapher struct {
	Prompt list.Prompt
	Store  profiles.Store
	Pack   cartogram.Pack
}

func (g *Grapher) Resolve(args, roleNames, profileNames []string) (Path, error) {
	account, err := g.selectTargetAccount(args)
	if err != nil {
		return Path{}, err
	}

	paths, err := g.findAllPaths(account)
	if err != nil {
		return Path{}, err
	}

	paths, err = g.filterByRole(paths, roleNames)
	if err != nil {
		return Path{}, err
	}

	paths, err = g.filterByProfile(paths, profileNames)
	if err != nil {
		return Path{}, err
	}

	if len(paths) > 1 {
		logger.InfoMsg("multiple valid paths detected. Selecting the first option")
	}
	return paths[0], nil
}

func (g *Grapher) selectTargetAccount(args []string) (cartogram.Account, error) {
	return g.Pack.FindWithPrompt(args, g.Prompt)
}

func (g *Grapher) findAllPaths(account cartogram.Account) ([]Path, error) {
	var allPaths []Path

	for _, r := range account.Roles {
		paths, err := g.findPathToRole(account, r)
		if err != nil {
			return []Path{}, err
		}
		allPaths = append(allPaths, paths...)
	}

	return allPaths, nil
}

func (g *Grapher) findPathToRole(account cartogram.Account, role cartogram.Role) ([]Path, error) {
	var allPaths []Path

	for _, item := range role.Sources {
		sourceMatch := roleSourceRegex.FindStringSubmatch(item.Path)
		if len(sourceMatch) == 3 {
			newAccount, newRole, ok := g.pathIsViable(sourceMatch[1], sourceMatch[2])
			if !ok {
				continue
			}
			newPaths, err := g.findPathToRole(newAccount, newRole)
			if err != nil {
				return []Path{}, err
			}
			allPaths = append(allPaths, newPaths...)
		} else {
			allPaths = append(allPaths, Path{{
				Profile: item.Path,
			}})
		}
	}

	myHop := Hop{
		Role:    role.Name,
		Account: account.Account,
		Mfa:     role.Mfa,
	}

	for i := range allPaths {
		allPaths[i] = append(allPaths[i], myHop)
	}
	return allPaths, nil
}

func (g *Grapher) pathIsViable(accountID, roleName string) (cartogram.Account, cartogram.Role, bool) {
	ok, account := g.Pack.Lookup(accountID)
	if !ok {
		logger.DebugMsgf("found dead end due to missing account: %s", accountID)
		return cartogram.Account{}, cartogram.Role{}, false
	}
	ok, role := account.Roles.Lookup(roleName)
	if !ok {
		logger.DebugMsgf(
			"Found dead end due to missing role: %s/%s", accountID, roleName,
		)
		return cartogram.Account{}, cartogram.Role{}, false
	}
	return account, role, true
}

func (g *Grapher) filterByRole(paths []Path, roleNames []string) ([]Path, error) {
	af := func(p Path) string { return p[len(p)-1].Role }
	allRoles := uniquePathAttributes(paths, af)
	role, err := list.WithInputSlice(
		g.Prompt,
		allRoles,
		roleNames,
		"Pick a role:",
	)
	if err != nil {
		return []Path{}, err
	}
	return filterPathsByAttribute(paths, role, af), nil
}

func (g *Grapher) filterByProfile(paths []Path, profileNames []string) ([]Path, error) {
	af := func(p Path) string { return p[0].Profile }

	allProfiles := uniquePathAttributes(paths, af)
	unionProfiles := sliceUnion(allProfiles, profileNames)

	profile, err := list.WithInputSlice(
		g.Prompt,
		allProfiles,
		unionProfiles,
		"Pick a profile:",
	)
	if err != nil {
		return []Path{}, err
	}

	return filterPathsByAttribute(paths, profile, af), nil
}
