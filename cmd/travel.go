package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/akerl/voyager/cartogram"

	"github.com/spf13/cobra"
)

const (
	accountRegexString = `(\d+)(/(\w+))?`
)

var accountRegex = regexp.MustCompile(accountRegexString)

var travelCmd = &cobra.Command{
	Use:   "travel",
	Short: "Resolve creds for a AWS account",
	RunE:  travelRunner,
}

func init() {
	rootCmd.AddCommand(travelCmd)
}

func travelRunner(cmd *cobra.Command, args []string) error {
	cp := cartogram.Pack{}
	if err := cp.Load(); err != nil {
		return err
	}

	if len(args) < 1 {
		return fmt.Errorf("Not enough arguments provided")
	}
	cName := args[0]
	c := cp[cName]
	if len(c) == 0 {
		return fmt.Errorf("Cartogram not found: %s", cName)
	}

	targetAccount, targetRole, err := findAccount(c, args[1:])
	if err != nil {
		return err
	}

	fmt.Printf("Account: %s\nRole: %s\n", targetAccount, targetRole)
	return nil
}

type filter struct {
	Name, Value string
}

func findAccount(c cartogram.Cartogram, args []string) (string, string, error) {
	var targetAccount, targetRole string

	targetAccount, targetRole, error = findDirectAccount(c, args)
	if err != nil {
		return "", "", err
	}
	if targetAccount != "" {
		return targetAccount, targetRole, nil
	}

	targetAccount, targetRole, err := parseMatchAccounts(c, args)
}

func findDirectAccount(c cartogram.Cartogram, args []string) (string, string, error) {
	var targetAccount, targetRole string
	if len(args) == 1 {
		accountMatch := accountRegex.FindStringSubmatch(args[0])
		if len(accountMatch) > 1 {
			targetAccount = accountMatch[1]
			if len(accountMatch) > 2 {
				targetRole = accountMatch[3]
			}
		}
	}

	return targetAccount, targetRole
}

func parseMatchAccounts(args []string) (string, string, error) {
	argPairs := parsePairs(args)
	matchingPack := cartogram.Pack{}
	for name, c := range cp {
		matchingCartogram := cartogram.Cartogram{}
		for _, account := range c {
			for _, f := range argPairs {
				for tagName, tagValue := range account.Tags {
					if tagName == f.Name || f.Name == "" {
						match, err := regexp.MatchString(f.Value, tagValue)
						if err != nil {
							return "", "", nil
						}
						if match {
							matchingCartogram = append(matchingCartogram, account)
						}
					}
				}
			}
		}
	}
}

func parseDirectAccount(args []string) (string, string) {
	var targetAccount, targetRole string
	if len(args) == 1 {
		accountMatch := accountRegex.FindStringSubmatch(args[0])
		if len(accountMatch) > 1 {
			targetAccount = accountMatch[1]
			if len(accountMatch) > 2 {
				targetRole = accountMatch[3]
			}
		}
	}
	return targetAccount, targetRole
}

func parsePairs(args []string) []filter {
	var argPairs []filter
	for _, a := range args {
		var f filter
		fields := strings.SplitN(a, ":", 2)
		if len(fields) == 1 {
			f.Value = fields[0]
		} else {
			f.Name = fields[0]
			f.Value = fields[1]
		}
		argPairs = append(argPairs, f)
	}
	return argPairs
}
