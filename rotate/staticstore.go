package rotate

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/iam"
)

type staticStore struct {
	Key iam.AccessKey
}

func (s *staticStore) Lookup(profile string) (credentials.Value, error) {
	logger.InfoMsgf(
		"Returning static values for profile %s: %s",
		profile,
		*s.Key.AccessKeyId,
	)
	return credentials.Value{
		AccessKeyID:     *s.Key.AccessKeyId,
		SecretAccessKey: *s.Key.SecretAccessKey,
	}, nil
}

func (s *staticStore) Check(_ string) bool {
	return true
}

func (s *staticStore) Delete(_ string) error {
	return nil
}
