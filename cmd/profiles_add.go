package cmd

import (
	"fmt"

	"github.com/akerl/voyager/v3/profiles"

	"github.com/akerl/input/list"
	"github.com/spf13/cobra"
)

func init() {
	profilesCmd.AddCommand(profilesAddCmd)
}

var profilesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "add new AWS credentials",
	RunE:  profilesAddRunner,
}

func profilesAddRunner(_ *cobra.Command, args []string) error {
	var inputProfile string
	if len(args) != 0 {
		inputProfile = args[0]
	}

	store := profiles.NewDefaultStore()

	allProfiles, err := getAllProfiles()
	if err != nil {
		return err
	}

	profile, err := list.WithInputString(
		list.Default(),
		allProfiles,
		inputProfile,
		"Profile to add",
	)
	if err != nil {
		return err
	}

	check := store.Check(profile)
	if check {
		fmt.Println(
			"Profile is already stored; if you wish to update it, use the rotate command. " +
				"If you want to remove it, use the remove command",
		)
		return nil
	}
	_, err = store.Lookup(profile)
	if err == nil {
		fmt.Println("Successfully added profile")
	}
	return err
}
