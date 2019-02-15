package prompt

// Args struct for prompt functions
type Args struct {
	Message string
	Options [][]string
}

// Func is the signature for prompts
type Func func(Args) (int, error)

// Types provides a map of prompt types by name
var Types = map[string]Func{
	"":      WithDefault,
	"wmenu": WithWmenu,
	"fuzzy": WithFuzzy,
}

// WithDefault uses the default prompt method
func WithDefault(a Args) (int, error) {
	return WithWmenu(a)
}
