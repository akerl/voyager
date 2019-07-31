package travel

import (
	"github.com/akerl/speculate/creds"
	"github.com/akerl/speculate/executors"
	"github.com/akerl/timber/log"

	"github.com/akerl/voyager/cartogram"
	"github.com/akerl/voyager/prompt"
)

var logger = log.NewLogger("voyager")

type hop struct {
	Profile string
	Account string
	Region  string
	Role    cartogram.Role
}

type voyage struct {
	pack    cartogram.Pack
	account cartogram.Account
	paths   [][]hop
	creds   creds.Creds
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
	return nil
}

func (v *voyage) loadCreds(i Itinerary) error {
	return nil
}
