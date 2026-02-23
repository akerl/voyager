package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/akerl/voyager/v3/pkgver"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("%s\n", pkgver.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
