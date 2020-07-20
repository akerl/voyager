package cmd

import (
	"github.com/akerl/voyager/v3/cartogram"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(profilesCmd)
}

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "manage stored AWS credentials",
}

func getAllProfiles() ([]string, error) {
	pack := cartogram.Pack{}
	if err := pack.Load(); err != nil {
		return []string{}, err
	}

	return pack.AllProfiles(), nil
}
