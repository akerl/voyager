package cartogram

import (
	"os"
	"os/user"
	"path"
)

const (
	configName  = ".cartograms"
	specVersion = 1
)

func configDir() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", err
	}
	dir := path.Join(home, configName)
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func homeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

type promptFunc func(string, []string, string) (string, error)
