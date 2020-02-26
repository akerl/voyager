package cmd

import (
	"fmt"

	"github.com/akerl/voyager/v2/profiles"

	"github.com/spf13/cobra"
)

func init() {
	profilesCmd.AddCommand(profilesShowCmd)
}

var profilesShowCmd = &cobra.Command{
	Use:   "show PROFILE",
	Short: "show a stored AWS credential",
	RunE:  profilesShowRunner,
}

func profilesShowRunner(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("no profile name provided")
	}
	profile := args[0]

	store := profiles.NewDefaultStore()

	check := store.Check(profile)
	if !check {
		return fmt.Errorf("no credentials stored for profile: %s", profile)
	}

	creds, err := store.Lookup(profile)
	if err != nil {
		return err
	}
	fmt.Printf("Access Key ID: %s\n", creds.AccessKeyID)
	fmt.Printf("Secret Access Key: %s\n", creds.SecretAccessKey)
	return nil
}
