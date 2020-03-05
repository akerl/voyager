package rotate

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/akerl/voyager/v2/cartogram"
	"github.com/akerl/voyager/v2/profiles"
	"github.com/akerl/voyager/v2/travel"
	"github.com/akerl/voyager/v2/utils"
	"github.com/akerl/voyager/v2/version"

	"github.com/akerl/input/list"
	"github.com/akerl/speculate/v2/creds"
	"github.com/akerl/timber/v2/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/mdp/qrterminal/v3"
	"github.com/pquerna/otp/totp"
)

var logger = log.NewLogger("voyager")

// Rotator is a helper for rotating credentials
type Rotator struct {
	Store          profiles.Store
	InputProfile   string
	MfaPrompt      creds.MfaPrompt
	profile        string
	originalCreds  credentials.Value
	validCreds     credentials.Value
	newCreds       credentials.Value
	existingMfaArn string
}

func (r *Rotator) getMfaPrompt() creds.MfaPrompt {
	if r.MfaPrompt == nil {
		r.MfaPrompt = &creds.DefaultMfaPrompt{}
	}
	return r.MfaPrompt
}

func (r *Rotator) getStore() profiles.Store {
	if r.Store == nil {
		r.Store = profiles.NewDefaultStore()
	}
	return r.Store
}

// Execute rotates the users keypair and MFA
func (r *Rotator) Execute() error { // revive:disable-line:cyclomatic
	err := utils.ConfirmText(
		"this is a breaking change",
		"This command makes the following changes:",
		"* Creates a new AWS access/secret keypair",
		"* Deletes your existing AWS access/secret keypair",
		"* Deletes any existing MFA device on your AWS user",
		"* Creates a new MFA device",
	)
	if err != nil {
		return err
	}

	profile, err := r.getProfile()
	if err != nil {
		return err
	}

	r.originalCreds, err = r.getStore().Lookup(profile)
	if err != nil {
		return err
	}
	r.validCreds = r.originalCreds

	if err := r.deleteOtherKey(); err != nil {
		return err
	}

	if err := r.swapForMfaSession(); err != nil {
		return err
	}

	if err := r.deleteMfaDevice(); err != nil {
		return err
	}

	if err := r.createMfaDevice(); err != nil {
		return err
	}

	if err := r.generateNewKey(); err != nil {
		return err
	}

	if err := r.waitForConsistency(); err != nil {
		return err
	}

	if err := r.testAuth(); err != nil {
		return err
	}

	if err := r.deleteOriginalKey(); err != nil {
		return err
	}

	fmt.Println("Credential rotation complete!")
	return nil
}

func (r *Rotator) getProfile() (string, error) {
	var err error

	if r.profile == "" {
		pack := cartogram.Pack{}
		if err := pack.Load(); err != nil {
			return "", err
		}

		allProfiles := pack.AllProfiles()
		p := list.WmenuPrompt{}

		r.profile, err = list.WithInputString(
			p,
			allProfiles,
			r.InputProfile,
			"Which team credentials would you like to rotate?",
		)
	}

	return r.profile, err
}

func (r *Rotator) getRegion() (string, error) {
	profile, err := r.getProfile()
	if err != nil {
		return "", err
	}
	var region string
	if strings.HasPrefix(profile, "gov_") {
		region = "us-gov-west-1"
	} else {
		region = "us-east-1"
	}
	logger.InfoMsgf("parsed region: %s", region)
	return region, nil
}

func (r *Rotator) getAwsSession() (*session.Session, error) {
	region, err := r.getRegion()
	if err != nil {
		return nil, err
	}
	if r.validCreds.AccessKeyID == "" {
		return nil, fmt.Errorf("no valid credentials set")
	}
	awsConfig := aws.NewConfig().WithRegion(region).WithCredentials(
		credentials.NewStaticCredentialsFromCreds(r.validCreds),
	)
	return session.NewSession(awsConfig)
}

func (r *Rotator) getIamClient() (*iam.IAM, error) {
	session, err := r.getAwsSession()
	if err != nil {
		return nil, err
	}
	logger.InfoMsg("loading new IAM client")
	return iam.New(session), nil
}

func (r *Rotator) getStsClient() (*sts.STS, error) {
	session, err := r.getAwsSession()
	if err != nil {
		return nil, err
	}
	logger.InfoMsg("loading new STS client")
	return sts.New(session), nil
}

func (r *Rotator) getUsername() (string, error) {
	stsClient, err := r.getStsClient()
	if err != nil {
		return "", err
	}
	logger.InfoMsg("looking up username")
	userRes, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	arnChunks := strings.Split(*userRes.Arn, "/")
	username := arnChunks[len(arnChunks)-1]
	logger.InfoMsgf("found username: %s", username)
	return username, nil
}

func (r *Rotator) swapForMfaSession() error { // revive:disable-line:cyclomatic
	iamClient, err := r.getIamClient()
	if err != nil {
		return err
	}

	region, err := r.getRegion()
	if err != nil {
		return err
	}

	logger.InfoMsg("list MFA devices")
	listMfaRes, err := iamClient.ListMFADevices(&iam.ListMFADevicesInput{})
	if err != nil {
		return err
	}
	listMfa := listMfaRes.MFADevices
	logger.InfoMsgf("found %d MFA devices", len(listMfa))
	if len(listMfa) == 0 {
		return nil
	}
	r.existingMfaArn = *listMfa[0].SerialNumber

	if r.validCreds.AccessKeyID == "" {
		return fmt.Errorf("no valid credentials set")
	}

	c := creds.Creds{
		AccessKey:    r.validCreds.AccessKeyID,
		SecretKey:    r.validCreds.SecretAccessKey,
		SessionToken: r.validCreds.SessionToken,
		Region:       region,
	}

	logger.InfoMsg("getting mfa session token")
	c, err = c.GetSessionToken(creds.GetSessionTokenOptions{
		UseMfa:    true,
		MfaPrompt: r.getMfaPrompt(),
	})
	if err != nil {
		return err
	}

	logger.InfoMsg("setting valid creds with MFA session")
	r.validCreds = credentials.Value{
		AccessKeyID:     c.AccessKey,
		SecretAccessKey: c.SecretKey,
		SessionToken:    c.SessionToken,
	}
	return nil
}

func (r *Rotator) deleteOtherKey() error {
	iamClient, err := r.getIamClient()
	if err != nil {
		return err
	}

	logger.InfoMsg("listing IAM keys")
	existingKeysRes, err := iamClient.ListAccessKeys(&iam.ListAccessKeysInput{})
	if err != nil {
		return err
	}
	existingKeys := existingKeysRes.AccessKeyMetadata
	if len(existingKeys) == 1 {
		return nil
	}

	err = utils.ConfirmText(
		"yes",
		"You already have 2 access keys on the account.",
		"Continuing will delete both of them.",
	)
	if err != nil {
		return err
	}

	for _, item := range existingKeys {
		if *item.AccessKeyId != r.originalCreds.AccessKeyID {
			logger.InfoMsgf("deleting other key: %s", *item.AccessKeyId)
			_, err := iamClient.DeleteAccessKey(
				&iam.DeleteAccessKeyInput{AccessKeyId: item.AccessKeyId},
			)
			return err
		}
	}
	return fmt.Errorf("other key vanished!?")
}

func (r *Rotator) deleteOriginalKey() error {
	iamClient, err := r.getIamClient()
	if err != nil {
		return err
	}

	logger.InfoMsgf(
		"deleting original key: %s",
		r.originalCreds.AccessKeyID,
	)
	_, err = iamClient.DeleteAccessKey(
		&iam.DeleteAccessKeyInput{AccessKeyId: &r.originalCreds.AccessKeyID},
	)
	return err
}

func (r *Rotator) generateNewKey() error {
	iamClient, err := r.getIamClient()
	if err != nil {
		return err
	}

	profile, err := r.getProfile()
	if err != nil {
		return err
	}

	logger.InfoMsg("creating new access key")
	newKeyRes, err := iamClient.CreateAccessKey(&iam.CreateAccessKeyInput{})
	if err != nil {
		return err
	}
	newKey := newKeyRes.AccessKey

	fmt.Println("New AWS key pair generated:")
	fmt.Printf("  Access Key ID: %s\n", *newKey.AccessKeyId)
	fmt.Printf("  Secret Access Key: %s\n", *newKey.SecretAccessKey)

	logger.InfoMsg("patching multistore with static creds")
	store := r.getStore()
	store.Delete(profile)
	writableStore, ok := store.(profiles.WritableStore)
	if !ok {
		return fmt.Errorf("provided store doesn't support writes")
	}
	err = writableStore.Write(
		profile,
		credentials.Value{
			AccessKeyID:     *newKey.AccessKeyId,
			SecretAccessKey: *newKey.SecretAccessKey,
		},
	)
	if err != nil {
		return err
	}
	c, err := store.Lookup(profile)
	if err != nil {
		return err
	}
	r.newCreds = c
	r.validCreds = c
	return nil
}

func (r *Rotator) deleteMfaDevice() error {
	if r.existingMfaArn == "" {
		return nil
	}

	iamClient, err := r.getIamClient()
	if err != nil {
		return err
	}

	username, err := r.getUsername()
	if err != nil {
		return err
	}

	logger.InfoMsg("deactivating MFA device")
	_, err = iamClient.DeactivateMFADevice(
		&iam.DeactivateMFADeviceInput{
			SerialNumber: &r.existingMfaArn,
			UserName:     &username,
		},
	)
	if err != nil {
		return err
	}

	logger.InfoMsg("deleting MFA device")
	_, err = iamClient.DeleteVirtualMFADevice(
		&iam.DeleteVirtualMFADeviceInput{
			SerialNumber: &r.existingMfaArn,
		},
	)
	return err
}

func (r *Rotator) createMfaDevice() error { // revive:disable-line:cyclomatic
	iamClient, err := r.getIamClient()
	if err != nil {
		return err
	}

	username, err := r.getUsername()
	if err != nil {
		return err
	}

	profile, err := r.getProfile()
	if err != nil {
		return err
	}

	logger.InfoMsg("creating MFA device")
	newMfaDeviceRes, err := iamClient.CreateVirtualMFADevice(
		&iam.CreateVirtualMFADeviceInput{VirtualMFADeviceName: &username},
	)
	if err != nil {
		return err
	}
	newMfaDevice := newMfaDeviceRes.VirtualMFADevice

	qrURL := fmt.Sprintf(
		"otpauth://totp/%s?secret=%s",
		profile,
		newMfaDevice.Base32StringSeed,
	)
	qrterminal.Generate(qrURL, qrterminal.L, os.Stdout)
	fmt.Printf("MFA secret: %s\n", newMfaDevice.Base32StringSeed)

	timeTwo := time.Now().Add(time.Second * -30)
	timeOne := timeTwo.Add(time.Second * -30)
	authCodeOne, err := totp.GenerateCode(string(newMfaDevice.Base32StringSeed), timeOne)
	if err != nil {
		return err
	}
	authCodeTwo, err := totp.GenerateCode(string(newMfaDevice.Base32StringSeed), timeTwo)
	if err != nil {
		return err
	}

	logger.InfoMsg("enabling MFA device")
	_, err = iamClient.EnableMFADevice(
		&iam.EnableMFADeviceInput{
			UserName:            &username,
			SerialNumber:        newMfaDevice.SerialNumber,
			AuthenticationCode1: &authCodeOne,
			AuthenticationCode2: &authCodeTwo,
		},
	)
	if err != nil {
		return err
	}

	logger.InfoMsg("attempting to store new mfa seed")
	mfaPrompt := r.getMfaPrompt()
	writer, ok := mfaPrompt.(creds.WritableMfaPrompt)
	if !ok {
		logger.InfoMsg("mfa prompt is not writable")
	}
	err = writer.Store(*newMfaDevice.SerialNumber, string(newMfaDevice.Base32StringSeed))
	if err != nil {
		fmt.Println("Failed to store new MFA seed, but treating as non-fatal.")
		fmt.Println("Please make sure to scan the above QR code.")
		fmt.Printf("Error message from storage attempt: %s", err)
		retryText := writer.RetryText(*newMfaDevice.SerialNumber)
		if retryText != "" {
			fmt.Println(retryText)
		}
	}
	return nil
}

func (r *Rotator) testAuth() error {
	profile, err := r.getProfile()
	if err != nil {
		return err
	}

	profileChunks := strings.SplitN(profile, "_", 2)
	partition := profileChunks[0]
	role := profileChunks[1]
	var account string
	if partition == "comm" {
		account = "^ops-auth$"
	} else {
		account = "^ops-auth-gov$"
	}

	username, err := r.getUsername()
	if err != nil {
		return err
	}

	mfaPrompt := r.getMfaPrompt()

	pack := cartogram.Pack{}
	if err := pack.Load(); err != nil {
		return err
	}

	grapher := travel.Grapher{Pack: pack}
	path, err := grapher.Resolve([]string{account}, []string{role}, []string{profile})
	if err != nil {
		return err
	}
	opts := travel.DefaultTraverseOptions()
	opts.UserAgentItems = []creds.UserAgentItem{{
		Name:    "voyager",
		Version: version.Version,
		Extra:   []string{"rotator"},
	}}
	opts.MfaPrompt = mfaPrompt
	opts.Store = r.Store
	opts.SessionName = username

	logger.InfoMsg("testing auth using new creds")
	c, err := path.TraverseWithOptions(opts)
	if err != nil {
		return err
	}

	openURL, err := c.ToCustomConsoleURL("")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "# %s\n", openURL)
	return nil
}

func (r *Rotator) waitForConsistency() error {
	fmt.Println("Waiting for new IAM keypair to sync across AWS's backend")
	fmt.Println("This may take up to a minute")
	for i := 0; i < 6; i++ {
		time.Sleep(5 * time.Second)
		stsClient, err := r.getStsClient()
		if err != nil {
			return err
		}
		_, err = stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err == nil {
			return nil
		}
		aerr, ok := err.(awserr.RequestFailure)
		if !ok || aerr.StatusCode() != 403 {
			return err
		}
		logger.InfoMsgf("pausing to retry after error: %s", aerr)
	}
	return fmt.Errorf("failed after 10 retries")
}
