package main

import (
	"os"
	"regexp"

	"github.com/akerl/voyager/v2/cartogram"
	"github.com/akerl/voyager/v2/profiles"

	"github.com/akerl/input/list"
	"github.com/akerl/speculate/v2/creds"
	"github.com/akerl/timber/v2/log"
)

type Cache struct {
	creds map[string]creds.Creds
}
