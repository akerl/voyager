package cartogram

import (
	"os"
	"os/user"
	"path"

	"github.com/akerl/timber/v2/log"
)

const (
	configName  = ".cartograms"
	specVersion = 2
)

var logger = log.NewLogger("voyager")

func configDir() (string, error) {
	logger.InfoMsg("looking up config dir")
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
	logger.InfoMsg("looking up home dir")
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}
