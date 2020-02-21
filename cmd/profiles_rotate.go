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

	utils.ConfirmText(
		"this is a breaking change",
		"This command makes the following changes:",
		"* Creates a new AWS access/secret keypair",
		"* Deletes your existing AWS access/secret keypair",
		"* Deletes any existing MFA device on your AWS user",
		"* Creates a new MFA device",
	)

	//r := rotate.Rotator{UseYubikey: useYubikey, Profile: inputProfile}
	//return r.Execute()
	// TODO: finish implementing
	return nil
}
