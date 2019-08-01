package cmd

import (
	"fmt"

	"github.com/akerl/voyager/prompt"
	"github.com/akerl/voyager/travel"
	"github.com/akerl/voyager/yubikey"
	"github.com/spf13/cobra"
)

var travelCmd = &cobra.Command{
	Use:   "travel",
	Short: "Resolve creds for a AWS account",
	RunE:  travelRunner,
}

func init() {
	rootCmd.AddCommand(travelCmd)
	travelCmd.Flags().StringP("role", "r", "", "Choose target role to use")
	travelCmd.Flags().String("profile", "", "Choose source profile to use")
	travelCmd.Flags().StringP("prompt", "p", "", "Choose prompt to use")
	travelCmd.Flags().BoolP("yubikey", "y", false, "Use Yubikey for MFA")
}

func travelRunner(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	flagRole, err := flags.GetString("role")
	if err != nil {
		return err
	}

	flagProfile, err := flags.GetString("profile")
	if err != nil {
		return err
	}

	promptFlag, err := flags.GetString("prompt")
	if err != nil {
		return err
	}
	promptFunc, ok := prompt.Types[promptFlag]
	if !ok {
		return fmt.Errorf("prompt type not found: %s", promptFlag)
	}

	useYubikey, err := flags.GetBool("yubikey")
	if err != nil {
		return err
	}

	i := travel.Itinerary{
		Args:        args,
		RoleName:    []string{flagRole},
		ProfileName: []string{flagProfile},
		Prompt:      promptFunc,
	}

	if useYubikey {
		i.MfaPrompt = yubikey.NewPrompt()
	}

	creds, err := i.Travel()
	if err != nil {
		return err
	}

	for _, line := range creds.ToEnvVars() {
		fmt.Println(line)
	}
	url, err := creds.ToConsoleURL()
	if err != nil {
		return err
	}
	fmt.Printf("# %s\n", url)

	return nil
}
