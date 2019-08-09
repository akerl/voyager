package travel

import (
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
	var paths [][]hop
	mapProfiles := make(map[string]bool)
	mapRoles := make(map[string]bool)

	account, err := i.getAccount()
	if err != nil {
		return []hop{}, err
	}

	for _, r := range account.Roles {
		p, err := i.tracePath(account, r)
		if err != nil {
			return []hop{}, err
		}
		for _, item := range p {
			paths = append(paths, item)
			mapRoles[item[len(item)-1].Role] = true
		}
	}

	allRoles := keys(mapRoles)
	role, err := list.WithInputSlice(i.getPrompt(), allRoles, i.RoleNames, "Pick a role:")
	if err != nil {
		return []hop{}, err
	}
	hopsWithMatchingRoles := [][]hop{}
	for _, item := range paths {
		if item[len(item)-1].Role == role {
			hopsWithMatchingRoles = append(hopsWithMatchingRoles, item)
			mapProfiles[item[0].Profile] = true
		}
	}

	allProfiles := keys(mapProfiles)
	unionProfiles := sliceUnion(allProfiles, i.ProfileNames)
	profile, err := list.WithInputSlice(
		i.getPrompt(),
		allProfiles,
		unionProfiles,
		"Pick a profile:",
	)
	if err != nil {
		return []hop{}, err
	}
	hopsWithMatchingProfiles := [][]hop{}
	for _, item := range hopsWithMatchingRoles {
		if item[0].Profile == profile {
			hopsWithMatchingProfiles = append(hopsWithMatchingProfiles, item)
		}
	}

	if len(hopsWithMatchingProfiles) > 1 {
		logger.InfoMsg("Multiple valid paths detected. Selecting the first option")
	}
	return hopsWithMatchingProfiles[0], nil
}

func (i *Itinerary) tracePath(acc cartogram.Account, role cartogram.Role) ([][]hop, error) {
	var srcHops [][]hop

	for _, item := range role.Sources {
		sourceMatch := roleSourceRegex.FindStringSubmatch(item.Path)
		if len(sourceMatch) == 3 {
			err := i.testSourcePath(sourceMatch[1], sourceMatch[2], &srcHops)
			if err != nil {
				return [][]hop{}, err
			}
		} else {
			srcHops = append(srcHops, []hop{{
				Profile: item.Path,
			}})
		}
	}

	myHop := hop{
		Role:    role.Name,
		Account: acc.Account,
		Region:  acc.Region,
		Mfa:     role.Mfa,
	}

	for i := range srcHops {
		srcHops[i] = append(srcHops[i], myHop)
	}
	return srcHops, nil
}

func (i *Itinerary) testSourcePath(srcAccID, srcRoleName string, srcHops *[][]hop) error {
	ok, srcAcc := i.pack.Lookup(srcAccID)
	if !ok {
		logger.DebugMsgf("Found dead end due to missing account: %s", srcAccID)
		return nil
	}
	ok, srcRole := srcAcc.Roles.Lookup(srcRoleName)
	if !ok {
		logger.DebugMsgf(
			"Found dead end due to missing role: %s/%s", srcAccID, srcRoleName,
		)
		return nil
	}
	newPaths, err := i.tracePath(srcAcc, srcRole)
	if err != nil {
		return err
	}
	for _, np := range newPaths {
		*srcHops = append(*srcHops, np)
	}
	return nil
}
