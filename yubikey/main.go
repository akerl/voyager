package yubikey

import (
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/akerl/timber/v2/log"
	"github.com/yawn/ykoath"
)

var logger = log.NewLogger("voyager")

const (
	configName  = ".voyager"
	mappingName = "yubikey"
)

type config struct {
	Mapping map[string]string
	Serials map[string]string
}

func mappingFile() (string, error) {
	logger.InfoMsg("looking up yubikey mapping file path")
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	mapping := path.Join(dir, mappingName)
	logger.InfoMsgf("resolved yubikey mapping file path: %s", mapping)
	return mapping, nil
}

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
	logger.InfoMsgf("resolved yubikey config dir: %s", dir)
	return dir, nil
}

func homeDir() (string, error) {
	logger.InfoMsg("looking up home dir")
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	logger.InfoMsgf("resolved yubikey home dir: %s", usr.HomeDir)
	return usr.HomeDir, nil
}

// Prompt defines a yubikey prompt object
type Prompt struct {
	mapping map[string]string
	serials map[string]string
}

// NewPrompt populates the yubikey mapping from a dotfile, if it exists
func NewPrompt() *Prompt {
	logger.InfoMsg("creating new yubikey prompt object")
	p := Prompt{}
	file, err := mappingFile()
	if err != nil {
		logger.InfoMsgf("failed to load mapping file: %s", err)
		return &p
	}
	err = p.AddMappingFromFile(file)
	if err != nil {
		logger.InfoMsgf("failed to load mapping: %s", err)
	}
	return &p
}

// AddMappingFromFile adds a mapping for OTP names from a file
func (p *Prompt) AddMappingFromFile(file string) error {
	logger.InfoMsgf("attempting to read mapping file: %s", file)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	c := config{}
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}
	p.AddMapping(c.Mapping)
	p.AddSerials(c.Serials)
	return nil
}

// AddMapping adds a mapping for OTP names
func (p *Prompt) AddMapping(mapping map[string]string) {
	logger.InfoMsgf("adding mapping: %+v", mapping)
	p.mapping = mapping
}

// AddSerials adds a serial lookup for OTP names
func (p *Prompt) AddSerials(serials map[string]string) {
	logger.InfoMsgf("adding serials: %+v", serials)
	p.serials = serials
}

// Prompt asks the yubikey for a code
func (p *Prompt) Prompt(arn string) (string, error) {
	logger.InfoMsgf("prompting for yubikey mfa for %s", arn)
	name := p.otpName(arn)
	exists := p.otpExists(name)
	if !exists {
		return "", fmt.Errorf("failed to connect to yubikey")
	}

	return p.otpCode(name)
}

// Store writes an OTP to the yubikey
func (p *Prompt) Store(arn, base32seed string) error {
	logger.InfoMsgf("storing mfa for %s", arn)
	name := p.otpName(arn)
	oath, err := p.getDevice(name)
	if err != nil {
		logger.InfoMsgf("failed to access yubikey: %s", err)
		return err
	}
	defer oath.Close()

	secret, err := base32.StdEncoding.DecodeString(base32seed)
	if err != nil {
		logger.InfoMsg("failed to decode Base32 seed")
		return err
	}
	return oath.Put(name, ykoath.HmacSha1, ykoath.Totp, 6, secret, true)
}

// RetryText returns helper text for retrying the MFA storage
func (p *Prompt) RetryText(arn string) string {
	otpName := p.otpName(arn)
	return fmt.Sprintf(
		"To retry, run the following command:\n"+
			"ykman oath add --oath-type TOTP -t %s\n"+
			"When prompted, enter the MFA secret key from above",
		otpName,
	)
}

// PluggedIn returns true if a yubikey device is found on the system
func (p *Prompt) PluggedIn() bool {
	oath, err := p.getDevice("")
	if err != nil {
		return false
	}
	oath.Close()
	return true
}

func (p *Prompt) otpName(arn string) string {
	if translated, ok := p.mapping[arn]; ok {
		logger.InfoMsgf("translating %s to %s", arn, translated)
		return translated
	}
	return arn
}

func (p *Prompt) otpExists(name string) bool {
	logger.InfoMsgf("checking for existing of %s", name)
	oath, err := p.getDevice(name)
	if err != nil {
		logger.InfoMsgf("failed to access yubikey: %s", err)
		return false
	}
	defer oath.Close()

	otps, err := oath.List()
	if err != nil {
		logger.InfoMsgf("failed to list yubikey: %s", err)
		return false
	}

	for _, otp := range otps {
		if otp.Name == name {
			logger.InfoMsgf("found matching OTP: %s", otp.Name)
			return true
		}

		logger.InfoMsgf("found non-matching OTP: %s", otp.Name)
	}
	return false
}

func (p *Prompt) otpCode(name string) (string, error) {
	logger.InfoMsgf("prompting for code for %s", name)
	oath, err := p.getDevice(name)
	if err != nil {
		logger.InfoMsgf("failed to access yubikey: %s", err)
		return "", err
	}
	defer oath.Close()

	return oath.Calculate(name, func(_ string) error {
		fmt.Fprintln(os.Stderr, "Touch the yubikey to use OTP")
		return nil
	})
}

func (p *Prompt) getDevice(name string) (*ykoath.OATH, error) {
	logger.InfoMsg("creating new yubikey oath device")
	oath, err := ykoath.NewFromSerial(p.serials[name])
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
