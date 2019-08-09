package cmd

import (
	"fmt"

	"github.com/akerl/voyager/v2/travel"
	"github.com/akerl/voyager/v2/yubikey"

	"github.com/akerl/input/list"
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
	travelCmd.Flags().String("service", "", "Service path for console URL")
}

// revive:disable-next-line:cyclomatic
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
	prompt, ok := list.Types[promptFlag]
	if !ok {
		return fmt.Errorf("prompt type not found: %s", promptFlag)
	}

	useYubikey, err := flags.GetBool("yubikey")
	if err != nil {
		return err
	}

	servicePath, err := flags.GetString("service")
	if err != nil {
		return err
	}

	i := travel.Itinerary{
		Args:        args,
		RoleName:    []string{flagRole},
		ProfileName: []string{flagProfile},
		Prompt:      prompt,
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
	url, err := creds.ToCustomConsoleURL(servicePath)
	if err != nil {
		return err
	}
	fmt.Printf("# %s\n", url)

	return nil
}
