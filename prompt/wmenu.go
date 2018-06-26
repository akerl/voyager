package prompt

import (
	"os"

	"github.com/dixonwille/wmenu"
)

// WithWmenu picks using wmenu
func WithWmenu(a Args) (string, error) {
	c := make(chan string, 1)

	// TODO: support picking by name
	menu := wmenu.NewMenu(a.Message)
	menu.ChangeReaderWriter(os.Stdin, os.Stderr, os.Stderr)
	menu.LoopOnInvalid()
	menu.Action(func(opts []wmenu.Opt) error {
		c <- opts[0].Value.(string)
		return nil
	})

	for _, item := range a.Options {
		isDefault := false
		if item == a.Default {
			isDefault = true
		}
		menu.Option(item, item, isDefault, nil)
	}

	if err := menu.Run(); err != nil {
		return "", err
	}

	return <-c, nil
}
