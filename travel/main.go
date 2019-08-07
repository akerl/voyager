package travel

import (
	"fmt"
	"os"

	"github.com/akerl/speculate/creds"
	"github.com/akerl/speculate/executors"
	"github.com/akerl/timber/log"

	"github.com/akerl/voyager/cartogram"
	"github.com/akerl/voyager/profiles"
	"github.com/akerl/voyager/prompt"
)

var logger = log.NewLogger("voyager")

type hop struct {
	Profile string
	Account string
	Region  string
	Role    string
	Mfa     bool
	Policy  string
}

// Itinerary describes a travel request
type Itinerary struct {
	Args        []string
	RoleName    []string
	ProfileName []string
	SessionName string
	Policy      string
	Lifetime    int64
	MfaCode     string
	MfaSerial   string
	MfaPrompt   executors.MfaPrompt
	Prompt      prompt.Func
	Store       profiles.Store
	pack        *cartogram.Pack
}

// Travel loads creds from a full set of parameters
func (i *Itinerary) Travel() (creds.Creds, error) {
	return i.getCreds()
}

func (i *Itinerary) getStore() profiles.Store {
	if i.Store == nil {
		logger.InfoMsg("Using default profile store")
		i.Store = profiles.NewDefaultStore()
	}
	return i.Store
}

func (i *Itinerary) getPrompt() prompt.Func {
	if i.Prompt == nil {
		logger.InfoMsg("Using default prompt")
		i.Prompt = prompt.WithDefault
	}
	return i.Prompt
}

func (i *Itinerary) getPack() (cartogram.Pack, error) {
	if i.pack != nil {
		return *i.pack, nil
	}
	i.pack = &cartogram.Pack{}
	err := i.pack.Load()
	return *i.pack, err
}

func (i *Itinerary) getAccount() (cartogram.Account, error) {
	pack, err := i.getPack()
	if err != nil {
		return cartogram.Account{}, err
	}
	return pack.FindWithPrompt(i.Args, i.getPrompt())
}

func (i *Itinerary) getCreds() (creds.Creds, error) {
	var c creds.Creds
	var err error

	path, err := i.getPath()
	if err != nil {
		return c, err
	}

	err = clearEnvironment()
	if err != nil {
		return c, err
	}

	profileHop, stack := path[0], path[1:]
	logger.InfoMsg(fmt.Sprintf("Setting origin hop: %+v", profileHop))
	store := i.getStore()
	err = profiles.SetProfile(profileHop.Profile, store)
	if err != nil {
		return c, err
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

func clearEnvironment() error {
	for varName := range creds.Translations["envvar"] {
		logger.InfoMsg(fmt.Sprintf("Unsetting env var: %s", varName))
		err := os.Unsetenv(varName)
		if err != nil {
			return err
		}
	}
	return nil
}

//revive:disable-next-line:cyclomatic
func (i *Itinerary) executeHop(thisHop hop, c creds.Creds) (creds.Creds, error) {
	var newCreds creds.Creds
	logger.InfoMsg(fmt.Sprintf("Executing hop: %+v", thisHop))
	a := executors.Assumption{}

	logger.InfoMsg(fmt.Sprintf("Setting AWS_DEFAULT_REGION to %s", thisHop.Region))
	err := os.Setenv("AWS_DEFAULT_REGION", thisHop.Region)
	if err != nil {
		return newCreds, err
	}

	if err := a.SetAccountID(thisHop.Account); err != nil {
		return newCreds, err
	}
	if err := a.SetRoleName(thisHop.Role); err != nil {
		return newCreds, err
	}
	if err := a.SetSessionName(i.SessionName); err != nil {
		return newCreds, err
	}
	if err := a.SetLifetime(i.Lifetime); err != nil {
		return newCreds, err
	}
	if err := a.SetPolicy(thisHop.Policy); err != nil {
		return newCreds, err
	}
	if thisHop.Mfa {
		if err := a.SetMfa(true); err != nil {
			return newCreds, err
		}
		if err := a.SetMfaSerial(i.MfaSerial); err != nil {
			return newCreds, err
		}
		if err := a.SetMfaCode(i.MfaCode); err != nil {
			return newCreds, err
		}
		if err := a.SetMfaPrompt(i.MfaPrompt); err != nil {
			return newCreds, err
		}
	}
	newCreds, err = a.ExecuteWithCreds(c)
	newCreds.Region = thisHop.Region
	return newCreds, err
}

func stringInSlice(list []string, key string) bool {
	for _, item := range list {
		if item == key {
			return true
		}
	}
	return false
}

func sliceUnion(a []string, b []string) []string {
	var res []string
	for _, item := range a {
		if stringInSlice(b, item) {
			res = append(res, item)
		}
	}
	return res
}

func (i *Itinerary) getPath() ([]hop, error) { //revive:disable-line:cyclomatic
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
	role, err := i.getPrompt().Filtered(i.RoleName, allRoles, "Desired target role:")
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
	unionProfiles := sliceUnion(allProfiles, i.ProfileName)
	profile, err := i.getPrompt().Filtered(unionProfiles, allProfiles, "Desired target profile:")
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
		pathMatch := cartogram.AccountRegex.FindStringSubmatch(item.Path)
		if len(pathMatch) == 3 {
			err := i.testSourcePath(pathMatch[1], pathMatch[2], &srcHops)
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
		logger.DebugMsg(fmt.Sprintf("Found dead end due to missing account: %s", srcAccID))
		return nil
	}
	ok, srcRole := srcAcc.Roles.Lookup(srcRoleName)
	if !ok {
		logger.DebugMsg(fmt.Sprintf(
			"Found dead end due to missing role: %s/%s", srcAccID, srcRoleName,
		))
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

func keys(input map[string]bool) []string {
	list := []string{}
	for k := range input {
		list = append(list, k)
	}
	return list
}
