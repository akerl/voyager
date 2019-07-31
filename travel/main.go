package travel

import (
	"fmt"

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
	pack             cartogram.Pack
	account          cartogram.Account
	paths            [][]hop
	creds            creds.Creds
	ProfileStoreName string
}

// Itinerary describes a travel request
type Itinerary struct {
	Args             []string
	RoleName         string
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
	v.ProfileStoreName = i.ProfileStoreName

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
	if err := v.loadPaths(i); err != nil {
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

func (v *voyage) loadPaths(i Itinerary) error {
	var paths [][]hop
	for _, r := range v.account.Roles {
		p, err := v.tracePath(v.account, r)
		if err != nil {
			return err
		}
		paths = append(paths, p)
	}
	fmt.Printf("%+v\n", paths)

	return nil
}

func (v *voyage) tracePath(acc cartogram.Account, role cartogram.Role) ([]hop, error) {
	var srcHops []hop
	var err error

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
				logger.DebugMsg(fmt.Sprintf("Found dead end due to missing role: %s/%s", srcAccID, srcRoleName))
				continue
			}
			srcHops, err = v.tracePath(srcAcc, srcRole)
			if err != nil {
				return []hop{}, err
			}
			if len(srcHops) != 0 {
				break
			}
		} else {
			store := profiles.Store{Name: v.ProfileStoreName}
			exists, err := store.CheckExists(item.Path)
			if err != nil {
				return []hop{}, err
			}
			if !exists {
				logger.DebugMsg(fmt.Sprintf("Found dead end due to missing credentials: %s", item.Path))
				continue
			}
			srcHops = []hop{{Profile: item.Path}}
			break
		}
	}

	srcHops = append(srcHops, hop{
		Role:    role.Name,
		Account: acc.Account,
		Region:  acc.Region,
		Mfa:     role.Mfa,
	})
	return srcHops, nil
}

func (v *voyage) loadCreds(i Itinerary) error {
	return nil
}
