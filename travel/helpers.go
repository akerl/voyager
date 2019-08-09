package travel

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

func keys(input map[string]bool) []string {
	list := []string{}
	for k := range input {
		list = append(list, k)
	}
	return list
}
