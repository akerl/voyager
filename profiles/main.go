package profiles

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/akerl/voyager/yubikey"

	"github.com/99designs/keyring"
	"github.com/akerl/timber/log"
)

var logger = log.NewLogger("voyager")

// Store holds a set of profiles
type Store struct {
	Name         string
	CustomLookup func(string, Store) (map[string]string, error)
}

func (s *Store) config() keyring.Config {
	return keyring.Config{
		AllowedBackends: []keyring.BackendType{
			"keychain",
			"wincred",
			"file",
		},
		KeychainName:             "login",
		KeychainTrustApplication: true,
		FilePasswordFunc:         filePasswordShim,
		FileDir:                  "~/.voyager/" + s.getName(),
		ServiceName:              "voyager:" + s.Name,
	}
}

func (s *Store) keyring() (keyring.Keyring, error) {
	return keyring.Open(s.config())
}

// SetProfile exports variables from keyring for the given profile
func (s *Store) SetProfile(profile string) error {
	if profile == "" {
		return fmt.Errorf("profile not set")
	}

	item, err := s.getItem(profile)
	if err != nil {
		return err
	}

	err = os.Setenv(yubikey.EnvVarName, profile)
	if err != nil {
		return err
	}
	for k, v := range item {
		logger.InfoMsg(fmt.Sprintf("Setting env var: %s", k))
		err = os.Setenv(k, v)
		if err != nil {
			return fmt.Errorf("failed to set env var: %s = %s", k, v)
		}
	}
	return nil
}

// CheckExists checks if a key exists in the keyring
func (s *Store) CheckExists(profile string) (bool, error) {
	if profile == "" {
		return false, fmt.Errorf("profile not set")
	}
	k, err := s.keyring()
	if err != nil {
		return false, err
	}
	itemName := s.itemName(profile)
	logger.InfoMsg(fmt.Sprintf("looking up in keyring: %s", itemName))
	_, err = k.Get(itemName)
	if err != nil {
		if err.Error() == keyring.ErrKeyNotFound.Error() {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

type itemStruct struct {
	EnvVars map[string]string
}

func (s *Store) getName() string {
	if s.Name == "" {
		logger.InfoMsg(fmt.Sprintf("set keyring store to default"))
		s.Name = "default"
	}
	return s.Name
}

func (s *Store) itemName(profile string) string {
	return fmt.Sprintf("voyager:%s:profile:%s", s.getName(), profile)
}

func (s *Store) parseItem(item keyring.Item) (map[string]string, error) {
	is := itemStruct{}
	err := json.Unmarshal(item.Data, &is)
	return is.EnvVars, err
}

func (s *Store) getItem(profile string) (map[string]string, error) {
	k, err := s.keyring()
	if err != nil {
		return map[string]string{}, err
	}
	itemName := s.itemName(profile)
	logger.InfoMsg(fmt.Sprintf("looking up in keyring: %s", itemName))
	item, err := k.Get(itemName)
	if err != nil && err.Error() == keyring.ErrKeyNotFound.Error() {
		logger.InfoMsg("falling back to env")
		return s.fallbackToCustom(profile)
	}
	return s.parseItem(item)
}

func (s *Store) fallbackToCustom(profile string) (map[string]string, error) {
	var err error
	var v map[string]string

	if s.CustomLookup != nil {
		logger.InfoMsg(fmt.Sprintf("looking up via custom lookup: %s", profile))
		v, err = s.CustomLookup(profile, *s)
	}
	if err != nil || s.CustomLookup == nil {
		logger.InfoMsg("falling back to prompt")
		v, err = s.fallbackToPrompt(profile)
		if err != nil {
			return map[string]string{}, err
		}
	}
	return s.migrateToStore(profile, v)
}

func (s *Store) fallbackToPrompt(profile string) (map[string]string, error) {
	fmt.Printf("Please enter your credentials for profile: %s\n", profile)
	accessKey, err := promptForInfo("AWS Access Key: ")
	if err != nil {
		return map[string]string{}, err
	}
	secretKey, err := promptForInfo("AWS Secret Key: ")
	if err != nil {
		return map[string]string{}, err
	}
	return map[string]string{
		"AWS_ACCESS_KEY_ID":     accessKey,
		"AWS_SECRET_ACCESS_KEY": secretKey,
	}, nil
}

func (s *Store) migrateToStore(profile string, value map[string]string) (map[string]string, error) {
	is := itemStruct{EnvVars: value}
	data, err := json.Marshal(is)
	if err != nil {
		return map[string]string{}, err
	}
	logger.InfoMsg("storing profile in keyring")
	k, err := s.keyring()
	if err != nil {
		return map[string]string{}, err
	}
	itemName := s.itemName(profile)
	err = k.Set(keyring.Item{
		Key:   itemName,
		Label: itemName,
		Data:  data,
	})
	if err != nil {
		return map[string]string{}, err
	}
	k, err = s.keyring()
	if err != nil {
		return map[string]string{}, err
	}
	logger.InfoMsg("looking up profile in keyring")
	item, err := k.Get(itemName)
	if err != nil {
		return map[string]string{}, err
	}
	return s.parseItem(item)
}

func promptForInfo(message string) (string, error) {
	infoReader := bufio.NewReader(os.Stdin)
	fmt.Fprint(os.Stderr, message)
	info, err := infoReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	info = strings.TrimRight(info, "\n")
	if info == "" {
		return "", fmt.Errorf("no input provided")
	}
	return info, nil
}

func filePasswordShim(_ string) (string, error) {
	return "", nil
}
