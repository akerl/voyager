package cmd

import (
	"fmt"
	"sort"

	"github.com/akerl/voyager/v2/profiles"

	"github.com/spf13/cobra"
)

func init() {
	profilesCmd.AddCommand(profilesListCmd)
}

var profilesListCmd = &cobra.Command{
	Use:   "list",
	Short: "list stored AWS credentials",
	RunE:  profilesListRunner,
}

func profilesListRunner(_ *cobra.Command, _ []string) error {
	allProfiles, err := getAllProfiles()
	if err != nil {
		return err
	}

	store := profiles.NewDefaultStore()
	existing := profiles.BulkCheck(store, allProfiles)

	if len(existing) == 0 {
		fmt.Println("No credentials found")
		return nil
	}

	sort.Strings(existing)
	for _, item := range existing {
		creds, err := store.Lookup(item)
		if err != nil {
			return err
		}
		fmt.Printf("%s (%s)\n", item, creds.AccessKeyID)
	}
	return nil
}
