package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// Session represents a shared AWS session
type Session struct {
	session *session.Session
}

// SessionConfig contains configuration for an AWS session
type SessionConfig struct {
	Region   string
	Endpoint string // For local development
}

// NewSession creates a new AWS session
func NewSession(config *SessionConfig) (*Session, error) {
	awsConfig := &aws.Config{
		Region: aws.String(config.Region),
	}

	// Use endpoint for local development
	if config.Endpoint != "" {
		awsConfig.Endpoint = aws.String(config.Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	// Create session
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	return &Session{
		session: sess,
	}, nil
}

// GetSession returns the underlying AWS session
func (s *Session) GetSession() *session.Session {
	return s.session
}
