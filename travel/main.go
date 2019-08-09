package travel

import (
	"os"
	"regexp"

	"github.com/akerl/voyager/v2/cartogram"
	"github.com/akerl/voyager/v2/profiles"

	"github.com/akerl/input/list"
	"github.com/akerl/speculate/v2/creds"
	"github.com/akerl/timber/v2/log"
)

var logger = log.NewLogger("voyager")

const (
	// roleSourceRegexString matches an account number and role name, /-delimited
	// Per https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_iam-limits.html .
	// role names can contain alphanumeric characters, and these symbols: +=,.@_-
	roleSourceRegexString = `^(\d{12})/([a-zA-Z0-9+=,.@_-]+)$`
)

var roleSourceRegex = regexp.MustCompile(roleSourceRegexString)

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
	Args         []string
	RoleNames    []string
	ProfileNames []string
	SessionName  string
	Policy       string
	Lifetime     int64
	MfaCode      string
	MfaSerial    string
	MfaPrompt    creds.MfaPrompt
	Prompt       list.Prompt
	Store        profiles.Store
	pack         cartogram.Pack
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

func (i *Itinerary) getPrompt() list.Prompt {
	if i.Prompt == nil {
		logger.InfoMsg("Using default prompt")
		i.Prompt = list.WmenuPrompt{}
	}
	return i.Prompt
}

func (i *Itinerary) getPack() (cartogram.Pack, error) {
	if i.pack != nil {
		return i.pack, nil
	}
	i.pack = cartogram.Pack{}
	err := i.pack.Load()
	return i.pack, err
}

func (i *Itinerary) getAccount() (cartogram.Account, error) {
	pack, err := i.getPack()
	if err != nil {
		return cartogram.Account{}, err
	}
	return pack.FindWithPrompt(i.Args, i.getPrompt())
}

func (i *Itinerary) getCreds() (creds.Creds, error) {
	path, err := i.getPath()
	if err != nil {
		return creds.Creds{}, err
	}

	err = clearEnvironment()
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

//revive:disable-next-line:cyclomatic
func (i *Itinerary) executeHop(thisHop hop, c creds.Creds) (creds.Creds, error) {
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
	profile, err := list.WithInputSlice(i.getPrompt(), allProfiles, unionProfiles, "Pick a profile:")
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

func keys(input map[string]bool) []string {
	list := []string{}
	for k := range input {
		list = append(list, k)
	}
	return list
}
