package yubikey

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/akerl/speculate/executors"
	"github.com/akerl/timber/log"
	"github.com/yawn/ykoath"
)

var logger = log.NewLogger("voyager")

const (
	configName  = ".voyager"
	mappingName = "yubikey"
)

type config struct {
	Mapping map[string]string
}

func mappingFile() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	mapping := path.Join(dir, mappingName)
	logger.InfoMsg(fmt.Sprintf("Resolved mapping file path: %s", mapping))
	return mapping, nil
}

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

// Prompt defines a yubikey prompt object
type Prompt struct {
	mapping  map[string]string
	Fallback bool
}

// NewPrompt populates the yubikey mapping from a dotfile, if it exists
func NewPrompt() *Prompt {
	p := Prompt{}
	file, err := mappingFile()
	if err != nil {
		logger.InfoMsg(fmt.Sprintf("Failed to load mapping file: %s", err))
		return &p
	}
	err = p.AddMappingFromFile(file)
	if err != nil {
		logger.InfoMsg(fmt.Sprintf("Failed to load mapping: %s", err))
	}
	return &p
}

// AddMappingFromFile adds a mapping for OTP names from a file
func (p *Prompt) AddMappingFromFile(file string) error {
	logger.DebugMsg(fmt.Sprintf("Reading mapping file: %s", file))
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	c := config{}
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}
	logger.DebugMsg(fmt.Sprintf("Parsed mapping file: %+v", c.Mapping))
	p.AddMapping(c.Mapping)
	return nil
}

// AddMappingFromFile adds a mapping for OTP names
func (p *Prompt) AddMapping(mapping map[string]string) {
	logger.DebugMsg(fmt.Sprintf("Adding mapping: %+v", mapping))
	p.mapping = mapping
}

// Prompt asks the yubikey for a code
func (p *Prompt) Prompt() (string, error) {
	name := p.otpName()
	exists := p.otpExists(name)
	if !exists {
		if p.Fallback {
			fallback := executors.DefaultMfaPrompt{}
			return fallback.Prompt()
		}
		return "", fmt.Errorf("Failed to connect to yubikey")
	}

	return p.otpCode(name)
}

func (p *Prompt) otpName() string {
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "default"
	}
	logger.InfoMsg(fmt.Sprintf("Yubikey prompt found AWS Profile: %s", profile))

	otpName := fmt.Sprintf("aws:%s", profile)
	if translated, ok := p.mapping[otpName]; ok {
		otpName = translated
	}
	logger.InfoMsg(fmt.Sprintf("Yubikey prompt using OTP name: %s", otpName))

	return otpName
}

func (p *Prompt) otpExists(name string) bool {
	oath, err := p.getDevice()
	if err != nil {
		logger.InfoMsg(fmt.Sprintf("Failed to access yubikey: %s", err))
		return false
	}
	defer oath.Close()

	otps, err := oath.List()
	if err != nil {
		logger.InfoMsg(fmt.Sprintf("Failed to list yubikey: %s", err))
	}

	for _, otp := range otps {
		if otp.Name == name {
			logger.InfoMsg(fmt.Sprintf("Found matching OTP: %s", otp.Name))
			return true
		}
		logger.InfoMsg(fmt.Sprintf("Found non-matching OTP: %s", otp.Name))
	}
	return false
}

func (p *Prompt) otpCode(name string) (string, error) {
	oath, err := p.getDevice()
	if err != nil {
		logger.InfoMsg(fmt.Sprintf("Failed to access yubikey: %s", err))
		return "", err
	}
	defer oath.Close()

	return oath.Calculate(name, func(_ string) error {
		fmt.Fprintln(os.Stderr, "Touch the yubikey to use OTP")
		return nil
	})
}

func (p *Prompt) getDevice() (*ykoath.OATH, error) {
	oath, err := ykoath.New()
	if err != nil {
		return nil, err
	}

	_, err = oath.Select()

	if err != nil {
		oath.Close()
		return nil, err
	}
	return oath, nil
}
