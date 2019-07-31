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
}

type voyage struct {
	pack    cartogram.Pack
	account cartogram.Account
	path    []hop
	creds   creds.Creds
}

// Itinerary describes a travel request
type Itinerary struct {
	Args             []string
	RoleName         string
	ProfileName      string
	SessionName      string
	Policy           string
	Lifetime         int64
	MfaCode          string
	MfaSerial        string
	MfaPrompt        executors.MfaPrompt
	Prompt           prompt.Func
	ProfileStoreName string
}

// Travel loads creds from a full set of parameters
func Travel(i Itinerary) (creds.Creds, error) {
	var c creds.Creds
	v := voyage{}

	if i.Prompt == nil {
		logger.InfoMsg("Using default prompt")
		i.Prompt = prompt.WithDefault
	}

	if err := v.loadPack(); err != nil {
		return c, err
	}
	if err := v.loadAccount(i.Args, i.Prompt); err != nil {
		return c, err
	}
	if err := v.loadPath(i); err != nil {
		return c, err
	}
	if err := v.loadCreds(i); err != nil {
		return c, err
	}
	return v.creds, nil
}

func (v *voyage) loadPack() error {
	v.pack = make(cartogram.Pack)
	return v.pack.Load()
}

func (v *voyage) loadAccount(args []string, pf prompt.Func) error {
	var err error
	v.account, err = v.pack.FindWithPrompt(args, pf)
	return err
}

func keys(input map[string]bool) []string {
	list := []string{}
	for k := range input {
		list = append(list, k)
	}
	return list
}

func (v *voyage) loadPath(i Itinerary) error {
	var paths [][]hop
	mapProfiles := make(map[string]bool)
	mapRoles := make(map[string]bool)

	for _, r := range v.account.Roles {
		p, err := v.tracePath(v.account, r)
		if err != nil {
			return err
		}
		for _, item := range p {
			paths = append(paths, item)
			mapRoles[item[len(item)-1].Role] = true
		}
	}

	allRoles := keys(mapRoles)
	role, err := i.Prompt.Simple(i.RoleName, allRoles, "Desired target role:")
	if err != nil {
		return err
	}
	hopsWithMatchingRoles := [][]hop{}
	for _, item := range paths {
		if item[len(item)-1].Role == role {
			hopsWithMatchingRoles = append(hopsWithMatchingRoles, item)
			mapProfiles[item[0].Profile] = true
		}
	}

	allProfiles := keys(mapProfiles)
	profile, err := i.Prompt.Simple(i.ProfileName, allProfiles, "Desired target profile:")
	if err != nil {
		return err
	}
	hopsWithMatchingProfiles := [][]hop{}
	for _, item := range paths {
		if item[0].Profile == profile {
			hopsWithMatchingProfiles = append(hopsWithMatchingProfiles, item)
		}
	}

	if len(hopsWithMatchingProfiles) > 1 {
		logger.InfoMsg("Multiple valid paths detected. Selecting the first option")
	}
	v.path = hopsWithMatchingProfiles[0]

	return nil
}

func (v *voyage) tracePath(acc cartogram.Account, role cartogram.Role) ([][]hop, error) {
	var srcHops [][]hop

	logger.DebugMsg(fmt.Sprintf("Tracing from %s / %s", acc.Account, role.Name))

	for _, item := range role.Sources {
		pathMatch := cartogram.AccountRegex.FindStringSubmatch(item.Path)
		if len(pathMatch) == 3 {
			srcAccID := pathMatch[1]
			ok, srcAcc := v.pack.Lookup(srcAccID)
			if !ok {
				logger.DebugMsg(fmt.Sprintf("Found dead end due to missing account: %s", srcAccID))
				continue
			}
			srcRoleName := pathMatch[2]
			ok, srcRole := srcAcc.Roles.Lookup(srcRoleName)
			if !ok {
				logger.DebugMsg(fmt.Sprintf(
					"Found dead end due to missing role: %s/%s", srcAccID, srcRoleName,
				))
				continue
			}
			newPaths, err := v.tracePath(srcAcc, srcRole)
			if err != nil {
				return srcHops, err
			}
			for _, np := range newPaths {
				srcHops = append(srcHops, np)
			}
		} else {
			//store := profiles.Store{Name: v.ProfileStoreName}
			//exists, err := store.CheckExists(item.Path)
			//if err != nil {
			//	return srcHops, err
			//}
			//if !exists {
			//	logger.DebugMsg(fmt.Sprintf(
			//      "Found dead end due to missing credentials: %s", item.Path,
			//  ))
			//	continue
			//}
			srcHops = append(srcHops, []hop{{Profile: item.Path}})
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

func (v *voyage) loadCreds(i Itinerary) error {
	var c creds.Creds
	var err error

	profileHop, stack := v.path[0], v.path[1:]
	for varName := range creds.Translations["envvar"] {
		logger.InfoMsg(fmt.Sprintf("Unsetting env var: %s", varName))
		err = os.Unsetenv(varName)
		if err != nil {
			return err
		}
	}
	store := profiles.Store{Name: i.ProfileStoreName}
	err = store.SetProfile(profileHop.Profile)
	if err != nil {
		return err
	}
	err = os.Setenv("AWS_DEFAULT_REGION", profileHop.Region)

	last := len(stack) - 1
	for index, thisHop := range stack {
		logger.InfoMsg(fmt.Sprintf("Executing hop: %+v", thisHop))
		a := executors.Assumption{}
		if err := a.SetAccountID(thisHop.Account); err != nil {
			return err
		}
		if err := a.SetRoleName(thisHop.Role); err != nil {
			return err
		}
		if err := a.SetSessionName(i.SessionName); err != nil {
			return err
		}
		if i.Lifetime != 0 {
			if err := a.SetLifetime(i.Lifetime); err != nil {
				return err
			}
		}
		if index == last {
			if err := a.SetPolicy(i.Policy); err != nil {
				return err
			}
		}
		if thisHop.Mfa {
			if err := a.SetMfa(true); err != nil {
				return err
			}
			if err := a.SetMfaSerial(i.MfaSerial); err != nil {
				return err
			}
			if err := a.SetMfaCode(i.MfaCode); err != nil {
				return err
			}
			if err := a.SetMfaPrompt(i.MfaPrompt); err != nil {
				return err
			}
		}
		c, err = a.ExecuteWithCreds(c)
		c.Region = thisHop.Region
		if err != nil {
			return err
		}
	}
	v.creds = c
	return nil
}
