package prompt

import (
	"os"
	"strings"

	"github.com/dixonwille/wmenu"
)

// WithWmenu picks using wmenu
func WithWmenu(a Args) (int, error) {
	c := make(chan int, 1)

	menu := wmenu.NewMenu(a.Message)
	menu.ChangeReaderWriter(os.Stdin, os.Stderr, os.Stderr)
	menu.LoopOnInvalid()
	menu.Action(func(opts []wmenu.Opt) error {
		c <- opts[0].ID
		return nil
	})

	for _, item := range a.Options {
		line := strings.Join(item, " ")
		menu.Option(line, nil, false, nil)
	}

	if err := menu.Run(); err != nil {
		return 0, err
	}

	return <-c, nil
}
