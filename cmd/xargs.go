package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/akerl/voyager/v2/cartogram"
	"github.com/akerl/voyager/v2/multi"
	"github.com/akerl/voyager/v2/travel"
	"github.com/akerl/voyager/v2/yubikey"

	"github.com/akerl/input/list"
	"github.com/akerl/speculate/v2/creds"
	"github.com/spf13/cobra"
)

var xargsCmd = &cobra.Command{
	Use:   "xargs",
	Short: "Run a command across many AWS accounts",
	RunE:  xargsRunner,
}

func init() {
	rootCmd.AddCommand(xargsCmd)
	xargsCmd.Flags().StringP("role", "r", "", "Choose target role to use")
	xargsCmd.Flags().String("profile", "", "Choose source profile to use")
	xargsCmd.Flags().StringP("prompt", "p", "", "Choose prompt to use")
	xargsCmd.Flags().BoolP("yubikey", "y", false, "Use Yubikey for MFA")
	xargsCmd.Flags().StringP("command", "c", "", "Command to execute")
}

// revive:disable-next-line:cyclomatic
func xargsRunner(cmd *cobra.Command, args []string) error {
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
	promptGenerator, ok := list.Types[promptFlag]
	if !ok {
		return fmt.Errorf("prompt type not found: %s", promptFlag)
	}
	prompt := promptGenerator()

	useYubikey, err := flags.GetBool("yubikey")
	if err != nil {
		return err
	}

	commandStr, err := flags.GetString("command")
	if err != nil {
		return err
	}
	if commandStr == "" {
		return fmt.Errorf("Command must be provided via --command / -c")
	}

	pack := cartogram.Pack{}
	if err := pack.Load(); err != nil {
		return err
	}

	grapher := travel.Grapher{
		Prompt: prompt,
		Pack:   pack,
	}

	opts := travel.DefaultTraverseOptions()
	if useYubikey {
		opts.MfaPrompt = &creds.MultiMfaPrompt{Backends: []creds.MfaPrompt{
			yubikey.NewPrompt(),
			&creds.DefaultMfaPrompt{},
		}}
	}

	processor := multi.Processor{
		Grapher:      grapher,
		Options:      opts,
		Args:         args,
		RoleNames:    []string{flagRole},
		ProfileNames: []string{flagProfile},
	}

	results, err := processor.ExecString(commandStr)
	if err != nil {
		return err
	}

	buffer, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(buffer))

	return nil
}
