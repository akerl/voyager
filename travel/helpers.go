package travel

import (
	"os"

	"github.com/akerl/speculate/v2/creds"
)

func clearEnvironment() error {
	for varName := range creds.Translations["envvar"] {
		logger.InfoMsgf("Unsetting env var: %s", varName)
		err := os.Unsetenv(varName)
		if err != nil {
			return err
		}
	}
	return nil
}

func stringInSlice(list []string, key string) bool {
	for _, item := range list {
		if item == key {
			return true
		}
	}
	return false
}

func sliceUnion(a []string, b []string) []string {
	var res []string
	for _, item := range a {
		if stringInSlice(b, item) {
			res = append(res, item)
		}
	}
	return res
}

type attrFunc func(Path) string

func uniquePathAttributes(paths []Path, af attrFunc) []string {
	tmpMap := map[string]bool{}
	for _, item := range paths {
		attr := af(item)
		tmpMap[attr] = true
	}
	tmpList := []string{}
	for item := range tmpMap {
		tmpList = append(tmpList, item)
	}
	return tmpList
}

func filterPathsByAttribute(paths []Path, match string, af attrFunc) []Path {
	filteredPaths := []Path{}
	for _, item := range paths {
		if af(item) == match {
			filteredPaths = append(filteredPaths, item)
		}
	}
	return filteredPaths
}
