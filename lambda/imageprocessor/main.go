package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Handler is our lambda handler invoked by the `lambda.Start` function call
func main() {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(getRegion()),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	// Create S3 service client
	s3Client := s3.New(sess)

	// Create image processor
	processor := NewImageProcessor(s3Client)

	// Start Lambda handler
	lambda.Start(processor.HandleRequest)
}

// getRegion gets the AWS region to use
func getRegion() string {
	// Default to us-east-1
	return "us-east-1"
}
