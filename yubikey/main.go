package yubikey

import (
	"fmt"
	"os"

	"github.com/akerl/speculate/executors"
	"github.com/akerl/timber/log"
	"github.com/yawn/ykoath"
)

var logger = log.NewLogger("voyager")

// Prompt defines a yubikey prompt object
type Prompt struct {
	mapping  map[string]string
	Fallback bool
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
	oath, err := ykoath.New()
	if err != nil {
		logger.InfoMsg(fmt.Sprintf("Failed to access yubikey: %s", err))
		return false
	}

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
	oath, err := ykoath.New()
	if err != nil {
		return "", err
	}
	return oath.Calculate(name, func(_ string) error {
		fmt.Fprintln(os.Stderr, "Touch the yubikey to use OTP")
		return nil
	})
}
