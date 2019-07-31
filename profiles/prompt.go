package profiles

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

// PromptStore is a storage backend which asks the user for input
type PromptStore struct{}

func (p *PromptStore) Lookup(profile string) (credentials.Value, error) {
	fmt.Printf("Please enter your credentials for profile: %s\n", profile)
	accessKey, err := p.getUserInput("AWS Access Key: ")
	if err != nil {
		return credentials.Value{}, err
	}
	secretKey, err := p.getUserInput("AWS Secret Key: ")
	if err != nil {
		return credentials.Value{}, err
	}
	return credentials.Value{
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
	}, nil
}

func (p *PromptStore) getUserInput(message string) (string, error) {
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
