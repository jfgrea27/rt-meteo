package blob

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3API interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type S3Store struct {
	bucket string
	client s3API
	log    *slog.Logger
}

func (s *S3Store) Save(key string, data []byte) error {
	s.log.Debug("saving to s3", "bucket", s.bucket, "key", key, "bytes", len(data))

	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		s.log.Error("failed to save to s3", "bucket", s.bucket, "key", key, "error", err)
		return fmt.Errorf("failed to put object s3://%s/%s: %w", s.bucket, key, err)
	}

	s.log.Info("saved to s3", "bucket", s.bucket, "key", key)
	return nil
}
