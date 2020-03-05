package profiles

import (
	"encoding/json"
	"fmt"

	"github.com/99designs/keyring"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// KeyringStore fetches credentials from the system keyring
type KeyringStore struct {
	Name string
}

// Lookup checks the keyring for credentials
func (k *KeyringStore) Lookup(profile string) (credentials.Value, error) {
	logger.InfoMsgf("looking up %s in keyring store", profile)
	ring, err := k.keyring()
	if err != nil {
		return credentials.Value{}, err
	}
	itemName := k.itemName(profile)
	logger.InfoMsgf("converted profile to item name: %s", itemName)
	item, err := ring.Get(itemName)
	if err != nil {
		return credentials.Value{}, err
	}
	return k.parseItem(item)
}

// Write caches the credentials for the user
func (k *KeyringStore) Write(profile string, creds credentials.Value) error {
	logger.InfoMsgf("writing %s in keyring store", profile)
	is := itemStruct{EnvVars: map[string]string{
		"AWS_ACCESS_KEY_ID":     creds.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY": creds.SecretAccessKey,
	}}
	data, err := json.Marshal(is)
	if err != nil {
		return err
	}
	ring, err := k.keyring()
	if err != nil {
		return err
	}
	itemName := k.itemName(profile)
	logger.InfoMsgf("converted profile to item name: %s", itemName)
	return ring.Set(keyring.Item{
		Key:   itemName,
		Label: itemName,
		Data:  data,
	})
}

// Check returns if the credentials are cached in the keyring
func (k *KeyringStore) Check(profile string) bool {
	logger.InfoMsgf("checking for %s in keyring store", profile)
	res, _ := k.Lookup(profile)
	return res.AccessKeyID != ""
}

// Delete removes a profile from the keyring
func (k *KeyringStore) Delete(profile string) error {
	logger.InfoMsgf("deleting for %s from keyring store", profile)
	ring, err := k.keyring()
	if err != nil {
		return err
	}
	itemName := k.itemName(profile)
	logger.InfoMsgf("converted profile to item name: %s", itemName)

	return ring.Remove(itemName)
}

func (k *KeyringStore) config() keyring.Config {
	return keyring.Config{
		AllowedBackends: []keyring.BackendType{
			"keychain",
			"wincred",
			"libsecret",
			"file",
		},
		KeychainName:            "login",
		LibSecretCollectionName: "voyager:" + k.getName(),
		FilePasswordFunc:        filePasswordShim,
		FileDir:                 "~/.voyager/" + k.getName(),
		ServiceName:             "voyager:" + k.getName(),
	}
}

func (k *KeyringStore) getName() string {
	if k.Name == "" {
		logger.InfoMsgf("set keyring store to default")
		k.Name = "default"
	}
	return k.Name
}

type itemStruct struct {
	EnvVars map[string]string
}

func (k *KeyringStore) parseItem(item keyring.Item) (credentials.Value, error) {
	is := itemStruct{}
	err := json.Unmarshal(item.Data, &is)
	if err != nil {
		return credentials.Value{}, err
	}
	return credentials.Value{
		AccessKeyID:     is.EnvVars["AWS_ACCESS_KEY_ID"],
		SecretAccessKey: is.EnvVars["AWS_SECRET_ACCESS_KEY"],
	}, nil
}

func (k *KeyringStore) itemName(profile string) string {
	return fmt.Sprintf("voyager:%s:profile:%s", k.getName(), profile)
}

func (k *KeyringStore) keyring() (keyring.Keyring, error) {
	return keyring.Open(k.config())
}

func filePasswordShim(_ string) (string, error) {
	return "", nil
}
