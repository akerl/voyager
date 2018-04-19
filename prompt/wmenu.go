package prompt

import (
	"os"

	"github.com/dixonwille/wmenu"
)

// PromptWmenu picks using wmenu
func PromptWmenu(message string, list []string, defaultOpt string) (string, error) {
	c := make(chan string, 1)

	// TODO: support picking by name
	menu := wmenu.NewMenu(message)
	menu.ChangeReaderWriter(os.Stdin, os.Stderr, os.Stderr)
	menu.LoopOnInvalid()
	menu.Action(func(opts []wmenu.Opt) error {
		c <- opts[0].Value.(string)
		return nil
	})

	for _, item := range list {
		isDefault := false
		if item == defaultOpt {
			isDefault = true
		}
		menu.Option(item, item, isDefault, nil)
	}

	if err := menu.Run(); err != nil {
		return "", err
	}

	return <-c, nil
}
