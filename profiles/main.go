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
	"github.com/aws/aws-sdk-go/aws/credentials"
)

var logger = log.NewLogger("voyager")

// Store holds a set of profiles
type Store struct {
	Name string
}

func (s *Store) config() keyring.Config {
	return keyring.Config{
		AllowedBackends: []keyring.BackendType{
			"keychain",
			"wincred",
			"file",
		},
		KeychainName:     "login",
		FilePasswordFunc: filePasswordShim,
		FileDir:          "~/.voyager/" + s.getName(),
		ServiceName:      "voyager:" + s.Name,
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
	variables, err := s.parseItem(item)
	for k, v := range variables {
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

func (s *Store) getItem(profile string) (keyring.Item, error) {
	k, err := s.keyring()
	if err != nil {
		return keyring.Item{}, err
	}
	itemName := s.itemName(profile)
	logger.InfoMsg(fmt.Sprintf("looking up in keyring: %s", itemName))
	item, err := k.Get(itemName)
	if err != nil && err.Error() == keyring.ErrKeyNotFound.Error() {
		logger.InfoMsg("falling back to env")
		return s.fallbackToEnv(profile)
	}
	return item, err
}

func (s *Store) fallbackToEnv(profile string) (keyring.Item, error) {
	logger.InfoMsg(fmt.Sprintf("looking up env: %s", profile))
	sc := credentials.NewSharedCredentials("", profile)
	v, err := sc.Get()
	if err != nil {
		logger.InfoMsg("falling back to prompt")
		v, err = s.fallbackToPrompt(profile)
		if err != nil {
			return keyring.Item{}, err
		}
	}
	return s.migrateToStore(profile, v)
}

func (s *Store) fallbackToPrompt(profile string) (credentials.Value, error) {
	fmt.Printf("Please enter your credentials for profile: %s\n", profile)
	accessKey, err := promptForInfo("AWS Access Key: ")
	if err != nil {
		return credentials.Value{}, err
	}
	secretKey, err := promptForInfo("AWS Secret Key: ")
	if err != nil {
		return credentials.Value{}, err
	}
	c := credentials.NewStaticCredentials(accessKey, secretKey, "")
	v, err := c.Get()
	return v, err
}

func (s *Store) migrateToStore(profile string, value credentials.Value) (keyring.Item, error) {
	is := itemStruct{
		EnvVars: map[string]string{
			"AWS_ACCESS_KEY_ID":     value.AccessKeyID,
			"AWS_SECRET_ACCESS_KEY": value.SecretAccessKey,
		},
	}
	data, err := json.Marshal(is)
	if err != nil {
		return keyring.Item{}, err
	}
	logger.InfoMsg("storing profile in keyring")
	k, err := s.keyring()
	if err != nil {
		return keyring.Item{}, err
	}
	itemName := s.itemName(profile)
	err = k.Set(keyring.Item{
		Key:   itemName,
		Label: itemName,
		Data:  data,
	})
	if err != nil {
		return keyring.Item{}, err
	}
	k, err = s.keyring()
	if err != nil {
		return keyring.Item{}, err
	}
	logger.InfoMsg("looking up profile in keyring")
	return k.Get(itemName)
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
