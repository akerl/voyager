package cartogram

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// TagFilter describes a filter to apply based on an account's tags
type TagFilter struct {
	Name  string
	Value *regexp.Regexp
}

// TagFilterSet describes a set of tag filters
type TagFilterSet []TagFilter

// Tags are a set of metadata about an account
type Tags map[string]string

// String converts tags to a human-readable string
func (t Tags) String() string {
	sortedTags := make([]string, len(t))
	i := 0
	for k := range t {
		sortedTags[i] = k
		i++
	}
	sort.Strings(sortedTags)

	var buffer bytes.Buffer
	for _, k := range sortedTags {
		buffer.WriteString(fmt.Sprintf("%s:%s, ", k, t[k]))
	}
	fullResult := buffer.String()
	return strings.TrimSuffix(fullResult, ", ")
}

// LoadFromArgs parses key:value args into a TagFilterSet
func (tfs *TagFilterSet) LoadFromArgs(args []string) error {
	var err error
	for _, a := range args {
		tf := TagFilter{}
		fields := strings.SplitN(a, ":", 2)
		var regexString string
		if len(fields) == 1 {
			regexString = fields[0]
		} else {
			tf.Name = fields[0]
			regexString = fields[1]
		}
		tf.Value, err = regexp.Compile(regexString)
		if err != nil {
			return err
		}
		*tfs = append(*tfs, tf)
	}
	return nil
}

// Match checks if an account matches the tag filter
func (tf TagFilter) Match(a Account) bool {
	if tf.Name != "" {
		if tf.Value.MatchString(a.Tags[tf.Name]) {
			return true
		}
	} else {
		for _, tagValue := range a.Tags {
			if tf.Value.MatchString(tagValue) {
				return true
			}
		}
	}
	return false
}

// Match checks if an account matches the tag filter set
func (tfs TagFilterSet) Match(a Account) bool {
	for _, tf := range tfs {
		if !tf.Match(a) {
			return false
		}
	}
	return true
}
