package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// CredentialConfig contains AWS credential configuration
type CredentialConfig struct {
	AccessKey  string
	SecretKey  string
	Profile    string
	UseProfile bool
}

// ConfigureSession configures AWS credentials for a session
func ConfigureSession(sess *session.Session, config *CredentialConfig) error {
	if config.UseProfile && config.Profile != "" {
		// Use a named profile
		sess.Config.Credentials = credentials.NewSharedCredentials("", config.Profile)
	} else if config.AccessKey != "" && config.SecretKey != "" {
		// Use explicit credentials
		sess.Config.Credentials = credentials.NewStaticCredentials(
			config.AccessKey,
			config.SecretKey,
			"", // Token - not needed for basic authentication
		)
	}
	// Otherwise, use the default credential chain

	return nil
}

// ConfigureLocalStack configures session for LocalStack (local AWS service emulator)
func ConfigureLocalStack(sess *session.Session, endpoint string) {
	sess.Config.Endpoint = aws.String(endpoint)
	sess.Config.S3ForcePathStyle = aws.Bool(true)
	sess.Config.DisableSSL = aws.Bool(true)
}
