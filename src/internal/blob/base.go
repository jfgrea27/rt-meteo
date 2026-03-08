package blob

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type BlobStore interface {
	Save(key string, data []byte) error
}

func ConstructS3Store(log *slog.Logger, bucket, awsRegion string, usePathStyle bool) *S3Store {
	if bucket == "" {
		panic("S3_BUCKET is required for s3 blob store")
	}
	if awsRegion == "" {
		panic("AWS_REGION is required for s3 blob store")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = usePathStyle
	})

	return &S3Store{
		bucket: bucket,
		client: client,
		log:    log.With("service", "s3", "bucket", bucket),
	}
}

func ConstructBlobStore(log *slog.Logger, provider string, bucket, awsRegion string, usePathStyle bool) BlobStore {
	switch provider {
	case "s3":
		return ConstructS3Store(log, bucket, awsRegion, usePathStyle)
	default:
		panic(fmt.Sprintf("%s is not a valid blob store provider", provider))
	}
}
