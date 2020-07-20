package cmd

import (
	"github.com/akerl/voyager/v3/rotate"
	"github.com/akerl/voyager/v3/yubikey"

	"github.com/akerl/speculate/v2/creds"
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

	var mfaPrompt creds.MfaPrompt
	if useYubikey {
		mfaPrompt = &creds.MultiMfaPrompt{Backends: []creds.MfaPrompt{
			yubikey.NewPrompt(),
			&creds.DefaultMfaPrompt{},
		}}
	} else {
		mfaPrompt = &creds.DefaultMfaPrompt{}
	}

	r := rotate.Rotator{
		InputProfile: inputProfile,
		MfaPrompt:    mfaPrompt,
	}
	return r.Execute()
}
