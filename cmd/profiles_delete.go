package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/akerl/voyager/v2/profiles"
	"github.com/akerl/voyager/v2/utils"

	"github.com/spf13/cobra"
)

func init() {
	profilesCmd.AddCommand(profilesDeleteCmd)
}

var profilesDeleteCmd = &cobra.Command{
	Use:   "delete PROFILE",
	Short: "delete a stored AWS credential",
	RunE:  profilesDeleteRunner,
}

func profilesDeleteRunner(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("no profile name provided")
	}
	profile := args[0]

	store := profiles.NewDefaultStore()

	check := store.Check(profile)
	if !check {
		fmt.Printf("No credentials stored for profile: %s\n", profile)
		return nil
	}

	err := utils.ConfirmText(
		"this is a destructive operation",
		fmt.Sprintf("This will delete the following profile: %s", profile),
	)
	if err != nil {
		return err
	}

	err = store.Delete(profile)
	if err == nil {
		fmt.Println("Deleted stored profile")
	}
	return err
}
