package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/akerl/voyager/v2/profiles"

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

	confirmText := "this is a destructive operation"
	fmt.Printf("This will delete the following profile: %s\n", profile)
	fmt.Printf("If you want to proceed, type '%s'\n", confirmText)

	confirmReader := bufio.NewReader(os.Stdin)
	confirmInput, err := confirmReader.ReadString('\n')
	if err != nil {
		return err
	}
	cleanedInput := strings.TrimSpace(confirmInput)
	if cleanedInput != confirmText {
		return fmt.Errorf("aborting")
	}

	err = store.Delete(profile)
	if err == nil {
		fmt.Println("Deleted stored profile")
	}
	return err
}
