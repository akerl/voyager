package cmd

import (
	"github.com/akerl/voyager/v2/rotate"

	"github.com/spf13/cobra"
)

func init() {
	profilesCmd.AddCommand(profilesRotateCmd)
	profilesRotateCmd.Flags().BoolP("yubikey", "y", false, "Store MFA on yubikey")
}

var profilesRotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "saves a new AWS keypair and MFA device from existing creds",
	RunE:  profilesRotateRunner,
}

func profilesRotateRunner(cmd *cobra.Command, args []string) error {
	var inputProfile string
	if len(args) != 0 {
		inputProfile = args[0]
	}

	useYubikey, err := cmd.Flags().GetBool("yubikey")
	if err != nil {
		return err
	}

	r := rotate.Rotator{
		UseYubikey:   useYubikey,
		InputProfile: inputProfile,
		Store:        profiles.NewDefaltStore(),
	}
	return r.Execute()
}
