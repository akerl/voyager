package cmd

import (
	"fmt"

	"github.com/akerl/voyager/prompt"
	"github.com/akerl/voyager/travel"

	"github.com/spf13/cobra"
)

var travelCmd = &cobra.Command{
	Use:   "travel",
	Short: "Resolve creds for a AWS account",
	RunE:  travelRunner,
}

func init() {
	rootCmd.AddCommand(travelCmd)
	travelCmd.Flags().StringP("role", "r", "", "Choose role to use")
	travelCmd.Flags().StringP("prompt", "p", "", "Choose prompt to use")
}

func travelRunner(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	flagRole, err := flags.GetString("role")
	if err != nil {
		return err
	}

	promptFlag, err := flags.GetString("prompt")
	if err != nil {
		return err
	}
	promptFunc, ok := prompt.Types[promptFlag]
	if !ok {
		return fmt.Errorf("Prompt type not found: %s", promptFlag)
	}

	i := travel.Itinerary{
		Args:     args,
		RoleName: flagRole,
		Prompt:   promptFunc,
	}
	creds, err := travel.Travel(i)
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
