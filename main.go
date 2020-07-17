package main

import (
	"os"

	"github.com/akerl/voyager/v3/cmd"

	"github.com/akerl/speculate/v2/helpers"
)

func main() {
	if err := cmd.Execute(); err != nil {
		helpers.PrintAwsError(err)
		os.Exit(1)
	}
}
