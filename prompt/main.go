package prompt

import (
	"fmt"
	"sort"
)

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

// Simple returns either the pre-provided selection or the result of a user prompt
func (f Func) Simple(val string, options []string, msg string) (string, error) {
	if val != "" {
		for _, item := range options {
			if item == val {
				return val, nil
			}
		}
		return "", fmt.Errorf("user provided selection not found: %s", val)
	}
	if len(options) == 1 {
		return options[0], nil
	}

	sort.Strings(options)

	slices := make([][]string, len(options))
	for index, item := range options {
		slices[index] = []string{item}
	}

	pa := Args{
		Message: msg,
		Options: slices,
	}
	index, err := f(pa)
	if err != nil {
		return "", err
	}

	return options[index], nil
}
