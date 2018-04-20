package prompt

// WithDefault uses the default prompt method
func WithDefault(message string, list []string, defaultOpt string) (string, error) {
	return WithWmenu(message, list, defaultOpt)
}
