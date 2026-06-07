package common

import (
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	s3Client   *s3.S3
	bucketName string
	s3Err      error
	s3Once     sync.Once
)

// GetS3Client initializes and returns the shared AWS S3 service client for Backblaze B2 in a thread-safe way
func GetS3Client() (*s3.S3, string, error) {
	s3Once.Do(func() {
		keyID := strings.TrimSpace(os.Getenv("B2_APPLICATION_KEY_ID"))
		appKey := strings.TrimSpace(os.Getenv("B2_APPLICATION_KEY"))
		endpoint := strings.TrimSpace(os.Getenv("B2_ENDPOINT"))
		bucketName = strings.TrimSpace(os.Getenv("B2_BUCKET_NAME"))

		if keyID == "" || appKey == "" || endpoint == "" || bucketName == "" {
			return
		}

		region := "us-east-1"
		endpointParts := strings.Split(endpoint, ".")
		if len(endpointParts) >= 2 {
			region = endpointParts[1]
		}

		s3Config := &aws.Config{
			Credentials:      credentials.NewStaticCredentials(keyID, appKey, ""),
			Endpoint:         aws.String("https://" + endpoint),
			Region:           aws.String(region),
			S3ForcePathStyle: aws.Bool(false),
		}

		sess, err := session.NewSession(s3Config)
		if err != nil {
			s3Err = err
			return
		}

		s3Client = s3.New(sess)
	})

	return s3Client, bucketName, s3Err
}
